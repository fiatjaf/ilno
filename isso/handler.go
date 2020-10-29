package isso

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fiatjaf/go-lnurl"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"wrong.wang/x/go-isso/extract"
	"wrong.wang/x/go-isso/response/json"
	"wrong.wang/x/go-isso/tool/bloomfilter"
	"wrong.wang/x/go-isso/tool/validator"
)

const maxlikeanddislikes = 142

// CreateComment create a new comment
func (isso *ISSO) CreateComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := RequestIDFromContext(r.Context())
		commentOrigin := FindOrigin(r)
		if commentOrigin == "" {
			json.BadRequest(requestID, w, nil, "can not find header origin")
			return
		}
		var comment submittedComment
		err := jsonBind(r.Body, &comment)
		if err != nil {
			json.BadRequest(requestID, w, err, descRequestInvalidParm)
			return
		}

		ok, err := lnurl.VerifySignature(comment.K1, comment.Sig, comment.Key)
		if !ok {
			json.Unauthorized(requestID, w, err,
				fmt.Sprintf("user credentials are invalid: %s", err.Error()))
			return
		}

		comment.URI = mux.Vars(r)["uri"]
		comment.RemoteAddr = findClientIP(r)
		if err := validator.Validate(comment); err != nil {
			json.BadRequest(requestID, w, err, fmt.Sprintf("comment validate failed: %s", err.Error()))
			return
		}
		if comment.Website != nil && (strings.HasPrefix(*comment.Website, "https://") || strings.HasPrefix(*comment.Website, "http://")) {
			*comment.Website = "http://" + *comment.Website
		}

		ok, reason := isso.newCommentGuard(r.Context(), comment.Comment, comment.URI)
		if !ok {
			json.Forbidden(requestID, w, nil, reason)
		}

		var thread Thread
		thread, err = isso.storage.GetThreadByURI(r.Context(), comment.URI)
		if err != nil {
			if errors.Is(err, ErrStorageNotFound) {
				// no thread realted to this uri
				// so create new thread
				if comment.Title == "" {
					ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
					defer cancel()
					comment.Title, comment.URI, err = extract.GetPageTitle(ctx, commentOrigin, comment.URI)
					if err != nil {
						json.NotFound(requestID, w, err, "URI does not exist or can parse or get title correctly")
						return
					}
				}
				if thread, err = isso.storage.NewThread(r.Context(), comment.URI, comment.Title); err != nil {
					json.ServerError(requestID, w, err, descStorageUnhandledError)
					return
				}
				isso.tools.event.Publish("comments.new:new-thread", thread)
			} else {
				// can not handled error
				json.ServerError(requestID, w, err, descStorageUnhandledError)
				return
			}
		}

		isso.tools.event.Publish("comments.new:before-save", thread)

		if isso.config.Moderation.Enable {
			if isso.config.Moderation.ApproveAcquaintance &&
				comment.Email != nil &&
				isso.storage.IsApprovedAuthor(r.Context(), *comment.Email) {
				comment.Mode = ModePublic
			} else {
				comment.Mode = ModeAccepted
			}
		} else {
			comment.Mode = ModePublic
		}
		c, err := isso.storage.NewComment(r.Context(), comment.Comment, thread.ID, comment.RemoteAddr)
		if err != nil {
			json.ServerError(requestID, w, err, descStorageUnhandledError)
			return
		}
		isso.tools.event.Publish("comments.new:after-save", thread, c)

		reply, _ := c.convert(false, isso.tools.hash, isso.tools.markdown)

		isso.tools.event.Publish("comments.new:finish", thread, c)

		isso.setcookie(c, w, false)

		if c.Mode == ModeAccepted {
			json.Accepted(w, reply)
		} else {
			json.Created(w, reply)
		}
	}
}

func (isso *ISSO) newCommentGuard(ctx context.Context, c Comment, uri string) (bool, string) {
	if !isso.config.Server.Guard.Enable {
		return true, ""
	}
	if isso.config.Server.Guard.RequireEmail && c.Email == nil {
		return false, "email address required but not provided"
	}
	if isso.config.Server.Guard.RequireAuthor && c.Author == "" {
		return false, "author address required but not provided"
	}

	g := isso.config.Server.Guard
	return isso.storage.NewCommentGuard(ctx, c, uri, g.RateLimit, g.DirectReply, g.ReplyToSelf, isso.config.MaxAge)
}

// FetchComments fetch all related comments
func (isso *ISSO) FetchComments() http.HandlerFunc {
	type urlParm struct {
		Parent      *int64  `schema:"parent"`
		Limit       int64   `schema:"limit"`
		NestedLimit int64   `schema:"nested_limit"`
		After       float64 `schema:"after"`
		Plain       int64   `schema:"plain"`
	}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	makeReplies := func(cs []Comment, after float64, limit int64, plain bool) []reply {
		var replies []reply
		var count int64
		if limit <= 0 {
			limit = int64(len(cs) + 1)
		}
		for _, c := range cs {
			if c.Created > after && count < limit {
				count++
				r, _ := c.convert(plain, isso.tools.hash, isso.tools.markdown)
				replies = append(replies, r)
			}
		}
		return replies
	}

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := RequestIDFromContext(r.Context())
		var urlparm urlParm
		err := decoder.Decode(&urlparm, r.URL.Query())
		if err != nil {
			json.BadRequest(requestID, w, err, descRequestInvalidParm)
			return
		}
		var parent int64
		if urlparm.Parent == nil {
			parent = -1
		} else {
			parent = *urlparm.Parent
		}
		var plain bool
		if urlparm.Plain != 0 {
			plain = true
		}

		replyCount, err := isso.storage.CountReply(r.Context(), mux.Vars(r)["uri"], ModePublic, urlparm.After)
		if err != nil {
			json.ServerError(requestID, w, err, descStorageUnhandledError)
			return
		}
		// param `after` may cause the loss of old comment's parent
		if _, ok := replyCount[parent]; !ok {
			replyCount[parent] = 0
		}

		commentsByParent, err := isso.storage.FetchCommentsByURI(r.Context(), mux.Vars(r)["uri"], parent, ModePublic, "id", true)
		if err != nil {
			json.ServerError(requestID, w, err, descStorageUnhandledError)
			return
		}
		rJSON := struct {
			TotalReplies  int64   `json:"total_replies"`
			Replies       []reply `json:"replies"`
			ID            *int64  `json:"id"`
			HiddenReplies int64   `json:"hidden_replies"`
		}{
			ID: urlparm.Parent,
		}

		// null parent, only fetch top-comment
		if parent == -1 {
			// parent == -1 means need all comment's, here TotalReplies means top-leval comments
			rJSON.TotalReplies = replyCount[0]

			rJSON.Replies = makeReplies(commentsByParent[0], urlparm.After, urlparm.Limit, plain)
			rJSON.HiddenReplies = rJSON.TotalReplies - int64(len(rJSON.Replies))
			var zero int64
			emptyarray := make([]reply, 0)
			for i := range rJSON.Replies {
				count, ok := replyCount[rJSON.Replies[i].ID]
				if !ok {
					rJSON.Replies[i].TotalReplies = &zero
					rJSON.Replies[i].Replies = &emptyarray
					rJSON.Replies[i].HiddenReplies = &zero
				} else {
					replies := makeReplies(commentsByParent[rJSON.Replies[i].ID], urlparm.After, urlparm.NestedLimit, plain)
					rJSON.Replies[i].TotalReplies = &count
					rJSON.Replies[i].Replies = &replies
					cc := *rJSON.Replies[i].TotalReplies - int64(len(*rJSON.Replies[i].Replies))
					rJSON.Replies[i].HiddenReplies = &cc
				}
			}

		} else if parent > 0 {
			rJSON.TotalReplies = replyCount[parent]
			rJSON.Replies = makeReplies(commentsByParent[parent], urlparm.After, urlparm.Limit, plain)
			rJSON.HiddenReplies = rJSON.TotalReplies - int64(len(rJSON.Replies))
		} else {
			// parent = 0 not exist
			rJSON.TotalReplies = 0
			rJSON.Replies = []reply{}
			rJSON.HiddenReplies = 0
		}
		json.OK(w, rJSON)
	}
}

// CountComment return every thread's comment amount
func (isso *ISSO) CountComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := RequestIDFromContext(r.Context())
		uris := []string{}
		err := jsonBind(r.Body, &uris)
		if err != nil {
			json.BadRequest(requestID, w, err, descRequestInvalidParm)
			return
		}

		counts := []int64{}
		if len(uris) == 0 {
			json.OK(w, counts)
			return
		}

		countsByURI, err := isso.storage.CountComment(r.Context(), uris)
		if err != nil {
			json.ServerError(requestID, w, err, descStorageUnhandledError)
			return
		}
		for _, i := range countsByURI {
			counts = append(counts, i)
		}
		json.OK(w, counts)
	}
}

// ViewComment return specific comment
func (isso *ISSO) ViewComment() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		requestID := RequestIDFromContext(req.Context())
		id, err := strconv.ParseInt(mux.Vars(req)["id"], 10, 64)
		if err != nil {
			json.BadRequest(requestID, w, err, descRequestInvalidParm)
			return
		}

		var plain bool
		if req.URL.Query().Get("plain") == "0" {
			plain = true
		}

		comment, err := isso.storage.GetComment(req.Context(), id)
		if err != nil {
			if errors.Is(err, ErrStorageNotFound) {
				json.NotFound(requestID, w, err, descStorageNotFound)
				return
			}
			json.ServerError(requestID, w, err, descStorageUnhandledError)
			return
		}

		r, _ := comment.convert(plain, isso.tools.hash, isso.tools.markdown)
		json.OK(w, r)
	}
}

// EditComment edit an existing comment.
// Editing a comment is only possible for a short period of time after it was created and only if the requestor has a valid cookie for it.
// Editing a comment will set a new edit cookie in the response.
func (isso *ISSO) EditComment() http.HandlerFunc {

	type editInput struct {
		Text    string  `json:"text"  validate:"required,gte=3,lte=65535"`
		Author  *string `json:"author"  validate:"omitempty,gte=1,lte=15"`
		Email   *string `json:"email"  validate:"omitempty,email"`
		Website *string `json:"website"  validate:"omitempty,url"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := RequestIDFromContext(r.Context())
		comment, ok := isso.checkcookies(w, r)
		if !ok {
			return
		}
		var ei editInput
		if err := jsonBind(r.Body, &ei); err != nil {
			json.BadRequest(requestID, w, err, descRequestInvalidParm)
			return
		}
		if err := validator.Validate(ei); err != nil {
			json.BadRequest(requestID, w, err, fmt.Sprintf("edit post data validate failed: %s", err.Error()))
			return
		}

		comment.Text = ei.Text
		if ei.Author != nil {
			comment.Author = *ei.Author
		}
		if ei.Email != nil {
			comment.Email = ei.Email
		}
		if ei.Website != nil {
			comment.Website = ei.Website
		}
		comment.Modified = new(float64)
		*comment.Modified = float64(time.Now().UnixNano()) / float64(1e9)

		c, err := isso.storage.EditComment(r.Context(), comment)
		if err != nil {
			json.ServerError(requestID, w, err, descStorageUnhandledError)
			return
		}

		isso.tools.event.Publish("comments.edit", c)

		reply, _ := c.convert(false, isso.tools.hash, isso.tools.markdown)
		isso.setcookie(c, w, false)
		json.OK(w, reply)
	}
}

// VoteComment used to like or dislike comment
func (isso *ISSO) VoteComment() http.HandlerFunc {
	type vresponse struct {
		Likes    int    `json:"likes"`
		Dislikes int    `json:"dislikes"`
		Msg      string `json:"message,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := RequestIDFromContext(r.Context())
		cid, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
		if err != nil {
			json.BadRequest(requestID, w, err, descRequestInvalidParm)
			return
		}
		var upvote bool
		if mux.Vars(r)["vote"] == "like" {
			upvote = true
		} else if mux.Vars(r)["vote"] == "dislike" {
			upvote = false
		} else {
			json.BadRequest(requestID, w, err, descRequestInvalidParm)
			return
		}

		c, err := isso.storage.GetComment(r.Context(), cid)
		if err != nil {
			if errors.Is(err, ErrStorageNotFound) {
				json.NotFound(requestID, w, err, descStorageNotFound)
				return
			}
			json.ServerError(requestID, w, err, descStorageUnhandledError)
			return
		}

		vr := vresponse{Likes: c.Likes, Dislikes: c.Dislikes}

		if c.Likes+c.Dislikes > maxlikeanddislikes {
			vr.Msg = fmt.Sprintf(`denied due to a "likes + dislikes" total too high (%d > %d)`,
				c.Likes+c.Dislikes, maxlikeanddislikes)
			json.OK(w, vr)
			return
		}

		remoteAddr := findClientIP(r)
		bf := bloomfilter.RecoverFrom(c.Voters, c.Likes+c.Dislikes)
		if bf.Contains([]byte(remoteAddr)) {
			vr.Msg = fmt.Sprintf(`denied because a vote has already been registered for this remote address: %s`, remoteAddr)
			json.OK(w, vr)
			return
		}
		bf.Add([]byte(remoteAddr))
		c.Voters = bf.Buffer()

		err = isso.storage.VoteComment(r.Context(), c, upvote)
		if err != nil {
			json.ServerError(requestID, w, err, descStorageUnhandledError)
			return
		}

		if upvote {
			vr.Likes++
		} else {
			vr.Dislikes++
		}
		json.OK(w, vr)
		return
	}
}

// DeleteComment delete a comment
func (isso *ISSO) DeleteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := RequestIDFromContext(r.Context())

		comment, ok := isso.checkcookies(w, r)
		if !ok {
			return
		}
		comment, err := isso.storage.DeleteComment(r.Context(), comment.ID)
		if err != nil {
			json.ServerError(requestID, w, err, descStorageUnhandledError)
			return
		}

		isso.tools.event.Publish("comments.delete", comment.ID)

		reply, _ := comment.convert(false, isso.tools.hash, isso.tools.markdown)
		isso.setcookie(comment, w, true)
		json.OK(w, reply)
	}
}

func (isso *ISSO) checkcookies(w http.ResponseWriter, r *http.Request) (Comment, bool) {
	requestID := RequestIDFromContext(r.Context())
	cid, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		json.BadRequest(requestID, w, err, descRequestInvalidParm)
		return Comment{}, false
	}

	var c Comment
	if cookie, err := r.Cookie(fmt.Sprintf("%v", cid)); err == nil {
		cvalue := make(map[int64][20]byte)
		if err = isso.tools.securecookie.Decode(fmt.Sprintf("%v", cid), cookie.Value, &cvalue); err == nil {
			if h, ok := cvalue[cid]; ok {
				if c, err = isso.storage.GetComment(r.Context(), cid); err == nil {
					if h == sha1.Sum([]byte(c.Text)) {
						return c, true
					}
				} else {
					if errors.Is(err, ErrStorageNotFound) {
						json.NotFound(requestID, w, err, descStorageNotFound)
						return Comment{}, false
					}
					json.ServerError(requestID, w, err, descStorageUnhandledError)
					return Comment{}, false
				}
			}
		}
	}
	json.Forbidden(requestID, w, err, descRequestInvalidCookies)
	return Comment{}, false
}

func (isso *ISSO) setcookie(c Comment, w http.ResponseWriter, delete bool) {
	if delete {
		cookie := &http.Cookie{
			Name:   fmt.Sprintf("%v", c.ID),
			Path:   "/",
			MaxAge: -1,
			Secure: true,
		}
		http.SetCookie(w, cookie)

		cookie = &http.Cookie{
			Name:   fmt.Sprintf("isso-%v", c.ID),
			Path:   "/",
			MaxAge: -1,
			Secure: true,
		}
		if v := cookie.String(); v != "" {
			w.Header().Add("X-Set-Cookie", v)
		}
		return
	}
	if encoded, err := isso.tools.securecookie.Encode(fmt.Sprintf("%v", c.ID),
		map[int64][20]byte{c.ID: sha1.Sum([]byte(c.Text))}); err == nil {
		cookie := &http.Cookie{
			Name:   fmt.Sprintf("%v", c.ID),
			Value:  encoded,
			Path:   "/",
			MaxAge: isso.config.MaxAge,
			Secure: true,
		}
		http.SetCookie(w, cookie)
		cookie = &http.Cookie{
			Name:   fmt.Sprintf("isso-%v", c.ID),
			Value:  encoded,
			Path:   "/",
			MaxAge: isso.config.MaxAge,
			Secure: true,
		}
		if v := cookie.String(); v != "" {
			w.Header().Add("X-Set-Cookie", v)
		}
	}
}
