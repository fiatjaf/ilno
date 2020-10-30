package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/fiatjaf/ilno/ilno"
	"github.com/fiatjaf/ilno/logger"
	"gopkg.in/guregu/null.v4"
)

// IsApprovedAuthor check if email has approved in 6 month
func (d *Database) IsApprovedAuthor(ctx context.Context, email string) bool {
	logger.Debug("email %s", email)
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()
	if email == "" {
		return false
	}
	var flag int64
	err := d.DB.QueryRowContext(ctx, d.statement["comment_is_previously_approved_author"], email).Scan(&flag)
	return (err == nil) && (flag == 1)
}

// NewComment add comment into database
func (d *Database) NewComment(ctx context.Context, c ilno.Comment, threadID int64) (ilno.Comment, error) {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()
	logger.Debug("create %s's comment at %d", c.Author, threadID)
	if c.Parent != nil {
		parent, err := d.getComment(ctx, *c.Parent)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ilno.Comment{}, wraperror(err)
			}
			return ilno.Comment{}, wraperror(err)
		}
		if parent.TID != threadID {
			return ilno.Comment{}, wraperror(err)
		}
		if parent.Parent.Valid {
			c.Parent = &parent.Parent.Int64
		}
	}

	nc := newNullComment(c, threadID)

	result, err := d.DB.ExecContext(ctx, d.statement["comment_new"],
		nc.TID, nc.Parent, nc.Created, nc.Modified, nc.Mode,
		nc.Text, nc.Key, nc.Author, nc.Voters)
	if err != nil {
		return ilno.Comment{}, wraperror(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return ilno.Comment{}, wraperror(err)
	}
	comment, err := d.GetComment(ctx, id)
	if err != nil {
		return ilno.Comment{}, wraperror(err)
	}
	return comment, nil
}

// GetComment get comment by ID
func (d *Database) GetComment(ctx context.Context, id int64) (ilno.Comment, error) {
	logger.Debug("get comment %d", id)
	nc, err := d.getComment(ctx, id)
	if err != nil {
		return ilno.Comment{}, wraperror(err)
	}
	return nc.ToComment(), nil
}

func (d *Database) getComment(ctx context.Context, id int64) (nullComment, error) {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()

	var nc nullComment
	voters := make([]byte, 256)
	err := d.DB.QueryRowContext(ctx, d.statement["comment_get_by_id"], id).Scan(
		&nc.TID, &nc.ID, &nc.Parent, &nc.Created, &nc.Modified, &nc.Mode,
		&nc.Text, &nc.Key, &nc.Author, &nc.Likes, &nc.Dislikes, &voters,
	)
	nc.Voters = voters
	if err != nil {
		return nc, err
	}
	return nc, nil
}

// CountReply return comment count for main thread's comment and all reply threads for one uri.
// 0 mean null parent
func (d *Database) CountReply(ctx context.Context, uri string, mode int) (map[int64]int64, error) {
	logger.Debug("uri: %s", uri)
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()

	counts := map[int64]int64{}

	rows, err := d.DB.QueryContext(ctx, d.statement["comment_count_reply"], uri, mode)
	if err != nil {
		return nil, wraperror(err)
	}
	defer rows.Close()
	for rows.Next() {
		var p null.Int
		var c int64
		err := rows.Scan(&p, &c)
		if err != nil {
			return nil, wraperror(err)
		}
		if p.Valid {
			counts[p.Int64] = c
		} else {
			counts[0] = c
		}
	}
	if rows.Err() != nil {
		return nil, wraperror(err)
	}
	return counts, nil
}

// FetchCommentsByURI fetch comments related uri with a lot of param
func (d *Database) FetchCommentsByURI(ctx context.Context, uri string, parent int64, mode int, orderBy string, asc bool) (map[int64][]ilno.Comment, error) {
	logger.Debug("uri: %s", uri)
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()

	switch orderBy {
	case "id", "created", "modified", "likes", "dislikes":
	default:
		orderBy = "id"
	}

	desc := ""
	if !asc {
		desc += ` DESC `
	}

	condition := fmt.Sprintf(" ORDER BY %s %s", orderBy, desc)

	var rows *sql.Rows
	var err error
	switch {
	case parent < 0:
		stmt := d.statement["comment_fetch_by_uri"] + condition
		rows, err = d.DB.QueryContext(ctx, stmt, uri, mode, mode)
	case parent == 0:
		stmt := d.statement["comment_fetch_by_uri"] + ` AND comments.parent IS NULL ` + condition
		rows, err = d.DB.QueryContext(ctx, stmt, uri, mode, mode)
	case parent > 0:
		stmt := d.statement["comment_fetch_by_uri"] + ` AND comments.parent=? ` + condition
		rows, err = d.DB.QueryContext(ctx, stmt, uri, mode, mode, parent)
	}

	defer rows.Close()
	if err != nil {
		return nil, wraperror(err)
	}

	commentsbyparent := map[int64][]ilno.Comment{}

	for rows.Next() {
		var nc nullComment

		err := rows.Scan(
			&nc.TID, &nc.ID, &nc.Parent, &nc.Created, &nc.Modified, &nc.Mode,
			&nc.Text, &nc.Key, &nc.Author, &nc.Likes,
			&nc.Dislikes, &nc.Voters,
		)
		if err != nil {
			return nil, wraperror(err)
		}
		if nc.Parent.Valid {
			commentsbyparent[nc.Parent.Int64] = append(commentsbyparent[nc.Parent.Int64], nc.ToComment())
		} else {
			commentsbyparent[0] = append(commentsbyparent[0], nc.ToComment())
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, wraperror(err)
	}
	return commentsbyparent, nil
}

// CountComment count comment per thread
func (d *Database) CountComment(ctx context.Context, uris []string) (map[string]int64, error) {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()
	logger.Debug("uris: %v", uris)
	commentByURI := map[string]int64{}
	if len(uris) == 0 {
		return commentByURI, nil
	}
	rows, err := d.DB.QueryContext(ctx, d.statement["comment_count"])
	defer rows.Close()

	if err != nil {
		return nil, wraperror(err)
	}

	for rows.Next() {
		var uri string
		var count int64
		err := rows.Scan(&uri, &count)
		if err != nil {
			return nil, wraperror(err)
		}
		commentByURI[uri] = count
	}

	err = rows.Err()
	if err != nil {
		return nil, wraperror(err)
	}

	uriMap := map[string]int64{}
	for _, uri := range uris {
		uriMap[uri] = commentByURI[uri]
	}
	return uriMap, nil
}

// ActivateComment Activate comment id if pending
func (d *Database) ActivateComment(ctx context.Context, id int64) error {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()
	logger.Debug("id: %d", id)

	var rowsaffected int64
	err := d.execstmt(ctx, &rowsaffected, nil, d.statement["comment_activate"], id)
	if err != nil {
		return wraperror(err)
	}
	if rowsaffected != 1 {
		return wraperror(ilno.ErrNotExpectAmount)
	}
	return nil
}

// EditComment edit comment
func (d *Database) EditComment(ctx context.Context, c ilno.Comment) (ilno.Comment, error) {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()
	logger.Debug("edit %s 's comment", c.Author)

	var rowsaffected int64
	if c.Modified == nil {
		return ilno.Comment{}, wraperror(ilno.ErrInvalidParam)
	}
	err := d.execstmt(ctx, &rowsaffected, nil, d.statement["comment_edit"],
		c.Text, c.Author, *c.Modified, c.ID)
	if err != nil {
		return ilno.Comment{}, wraperror(err)
	}
	if rowsaffected != 1 {
		return ilno.Comment{}, wraperror(ilno.ErrNotExpectAmount)
	}

	comment, err := d.GetComment(ctx, c.ID)
	if err != nil {
		return ilno.Comment{}, wraperror(err)
	}
	return comment, nil
}

// DeleteComment delete comment by id
func (d *Database) DeleteComment(ctx context.Context, cid int64) (ilno.Comment, error) {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()
	logger.Debug("delete comment %d", cid)

	var n int64
	var err error
	if err = d.DB.QueryRowContext(ctx, d.statement["comment_delete_check"], cid).Scan(&n); err == nil {
		stmt := d.statement["comment_delete_hard"]
		if n > 0 {
			stmt = d.statement["comment_delete_soft"]
		}
		if err = d.execstmt(ctx, nil, nil, stmt, cid); err == nil {
			if err = d.execstmt(ctx, nil, nil, d.statement["comment_delete_stale"]); err == nil {
				if n > 0 {
					return d.GetComment(ctx, cid)
				}
				return ilno.Comment{}, nil
			}
		}
	}
	return ilno.Comment{}, wraperror(err)
}

// VoteComment vote  comment, but if may failed when break limit
func (d *Database) VoteComment(ctx context.Context, c ilno.Comment, up bool) error {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()
	logger.Debug("vote comment %d", c.ID)

	if up {
		c.Likes++
	} else {
		c.Dislikes++
	}

	voters := make([]byte, 256)
	copy(voters, c.Voters[:])
	err := d.execstmt(ctx, nil, nil, d.statement["comment_vote_set"], c.Likes, c.Dislikes, voters, c.ID)
	if err != nil {
		return wraperror(err)
	}
	return nil
}
