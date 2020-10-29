package ilno

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/fiatjaf/go-lnurl"
	"github.com/fiatjaf/ilno/extract"
	"github.com/fiatjaf/ilno/response/json"
	"github.com/fiatjaf/ilno/tool/bloomfilter"
	"github.com/fiatjaf/ilno/tool/validator"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

const maxlikeanddislikes = 142

// CreateComment create a new comment
func (ilno *ILNO) CreateComment() http.HandlerFunc {
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
		comment.Key = comment.Key

		var thread Thread
		thread, err = ilno.storage.GetThreadByURI(r.Context(), comment.URI)
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
				if thread, err = ilno.storage.NewThread(r.Context(), comment.URI, comment.Title); err != nil {
					json.ServerError(requestID, w, err, descStorageUnhandledError)
					return
				}
				ilno.tools.event.Publish("comments.new:new-thread", thread)
			} else {
				// can not handled error
				json.ServerError(requestID, w, err, descStorageUnhandledError)
				return
			}
		}

		ilno.tools.event.Publish("comments.new:before-save", thread)

		comment.Mode = ModePublic

		c, err := ilno.storage.NewComment(r.Context(), comment.Comment, thread.ID)
		if err != nil {
			json.ServerError(requestID, w, err, descStorageUnhandledError)
			return
		}
		ilno.tools.event.Publish("comments.new:after-save", thread, c)

		reply := reply{Comment: c}

		ilno.tools.event.Publish("comments.new:finish", thread, c)

		if c.Mode == ModeAccepted {
			json.Accepted(w, reply)
		} else {
			json.Created(w, reply)
		}
	}
}

// FetchComments fetch all related comments
func (ilno *ILNO) FetchComments() http.HandlerFunc {
	type urlParm struct {
		Parent *int64 `schema:"parent"`
	}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	makeReplies := func(cs []Comment) []reply {
		var replies []reply
		var count int64
		for _, c := range cs {
			count++
			reply := reply{Comment: c}
			replies = append(replies, reply)
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

		replyCount, err := ilno.storage.CountReply(r.Context(), mux.Vars(r)["uri"],
			ModePublic)
		if err != nil {
			json.ServerError(requestID, w, err, descStorageUnhandledError)
			return
		}
		// param `after` may cause the loss of old comment's parent
		if _, ok := replyCount[parent]; !ok {
			replyCount[parent] = 0
		}

		commentsByParent, err := ilno.storage.FetchCommentsByURI(r.Context(),
			mux.Vars(r)["uri"], parent, ModePublic, "id", true)
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

			rJSON.Replies = makeReplies(commentsByParent[0])
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
					replies := makeReplies(commentsByParent[rJSON.Replies[i].ID])
					rJSON.Replies[i].TotalReplies = &count
					rJSON.Replies[i].Replies = &replies
					cc := *rJSON.Replies[i].TotalReplies - int64(len(*rJSON.Replies[i].Replies))
					rJSON.Replies[i].HiddenReplies = &cc
				}
			}

		} else if parent > 0 {
			rJSON.TotalReplies = replyCount[parent]
			rJSON.Replies = makeReplies(commentsByParent[parent])
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
func (ilno *ILNO) CountComment() http.HandlerFunc {
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

		countsByURI, err := ilno.storage.CountComment(r.Context(), uris)
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
func (ilno *ILNO) ViewComment() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		requestID := RequestIDFromContext(req.Context())
		id, err := strconv.ParseInt(mux.Vars(req)["id"], 10, 64)
		if err != nil {
			json.BadRequest(requestID, w, err, descRequestInvalidParm)
			return
		}

		comment, err := ilno.storage.GetComment(req.Context(), id)
		if err != nil {
			if errors.Is(err, ErrStorageNotFound) {
				json.NotFound(requestID, w, err, descStorageNotFound)
				return
			}
			json.ServerError(requestID, w, err, descStorageUnhandledError)
			return
		}

		reply := reply{Comment: comment}
		json.OK(w, reply)
	}
}

// EditComment edit an existing comment.
// Editing a comment is only possible for a short period of time after it was created and only if the requestor has a valid lnurl-auth key for it.
func (ilno *ILNO) EditComment() http.HandlerFunc {
	type editInput struct {
		Text   string  `json:"text"  validate:"required,gte=3,lte=65535"`
		Author *string `json:"author"  validate:"omitempty,gte=1,lte=15"`
		Key    string  `json:"key"`
		Sig    string  `json:"sig"`
		K1     string  `json:"k1"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := RequestIDFromContext(r.Context())
		cid, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
		if err != nil {
			json.BadRequest(requestID, w, err, descRequestInvalidParm)
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

		comment, err := ilno.storage.GetComment(r.Context(), cid)
		if err != nil {
			if errors.Is(err, ErrStorageNotFound) {
				json.NotFound(requestID, w, err, descStorageNotFound)
				return
			}
			json.ServerError(requestID, w, err, descStorageUnhandledError)
			return
		}

		if comment.Key != ei.Key {
			json.Unauthorized(requestID, w, err, fmt.Sprintf("key doesn't match"))
			return
		}

		ok, err := lnurl.VerifySignature(ei.K1, ei.Sig, ei.Key)
		if !ok {
			json.Unauthorized(requestID, w, err, fmt.Sprintf("signature doesn't match: %s", err.Error()))
			return
		}

		comment.Text = ei.Text
		if ei.Author != nil {
			comment.Author = *ei.Author
		}
		comment.Modified = new(float64)
		*comment.Modified = float64(time.Now().UnixNano()) / float64(1e9)

		c, err := ilno.storage.EditComment(r.Context(), comment)
		if err != nil {
			json.ServerError(requestID, w, err, descStorageUnhandledError)
			return
		}

		ilno.tools.event.Publish("comments.edit", c)

		reply := reply{Comment: c}
		json.OK(w, reply)
	}
}

// VoteComment used to like or dislike comment
func (ilno *ILNO) VoteComment() http.HandlerFunc {
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

		c, err := ilno.storage.GetComment(r.Context(), cid)
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

		err = ilno.storage.VoteComment(r.Context(), c, upvote)
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
func (ilno *ILNO) DeleteComment() http.HandlerFunc {
	type deleteInput struct {
		Key string `json:"key"`
		Sig string `json:"sig"`
		K1  string `json:"k1"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := RequestIDFromContext(r.Context())
		cid, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
		if err != nil {
			json.BadRequest(requestID, w, err, descRequestInvalidParm)
			return
		}

		comment, err := ilno.storage.GetComment(r.Context(), cid)
		if err != nil {
			if errors.Is(err, ErrStorageNotFound) {
				json.NotFound(requestID, w, err, descStorageNotFound)
				return
			}
			json.ServerError(requestID, w, err, descStorageUnhandledError)
			return
		}

		var ei deleteInput
		if err := jsonBind(r.Body, &ei); err != nil {
			json.BadRequest(requestID, w, err, descRequestInvalidParm)
			return
		}

		if comment.Key != ei.Key {
			json.Unauthorized(requestID, w, err, fmt.Sprintf("key doesn't match"))
			return
		}

		ok, err := lnurl.VerifySignature(ei.K1, ei.Sig, ei.Key)
		if !ok {
			json.Unauthorized(requestID, w, err, fmt.Sprintf("signature doesn't match: %s", err.Error()))
			return
		}

		comment, err = ilno.storage.DeleteComment(r.Context(), comment.ID)
		if err != nil {
			json.ServerError(requestID, w, err, descStorageUnhandledError)
			return
		}

		ilno.tools.event.Publish("comments.delete", comment.ID)

		reply := reply{Comment: comment}
		json.OK(w, reply)
	}
}
