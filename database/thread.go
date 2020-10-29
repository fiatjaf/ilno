package database

import (
	"context"

	"github.com/fiatjaf/ilno/ilno"
	"github.com/fiatjaf/ilno/logger"
)

// GetThreadByURI get thread by uri
func (d *Database) GetThreadByURI(ctx context.Context, uri string) (ilno.Thread, error) {
	logger.Debug("uri %s", uri)
	var thread ilno.Thread
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()
	err := d.DB.QueryRowContext(ctx, d.statement["thread_get_by_uri"], uri).Scan(&thread.ID, &thread.URI, &thread.Title)
	if err != nil {
		return thread, wraperror(err)
	}
	return thread, nil
}

// GetThreadByID get thread by id
func (d *Database) GetThreadByID(ctx context.Context, id int64) (ilno.Thread, error) {
	logger.Debug("id %d", id)
	var thread ilno.Thread
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()
	err := d.DB.QueryRowContext(ctx, d.statement["thread_get_by_id"], id).Scan(&thread.ID, &thread.URI, &thread.Title)
	if err != nil {
		return thread, wraperror(err)
	}
	return thread, nil
}

// NewThread new a thread
func (d *Database) NewThread(ctx context.Context, uri string, title string) (ilno.Thread, error) {
	logger.Debug("create thread %s %s", uri, title)
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()

	if title == "" || uri == "" {
		return ilno.Thread{}, wraperror(ilno.ErrInvalidParam)
	}

	var rowsaffected, lastinsertid int64
	err := d.execstmt(ctx, &rowsaffected, &lastinsertid, d.statement["thread_new"], uri, title)
	if err != nil {
		return ilno.Thread{}, wraperror(err)
	}
	if rowsaffected != 1 {
		return ilno.Thread{}, wraperror(ilno.ErrNotExpectAmount)
	}
	return ilno.Thread{ID: lastinsertid, URI: uri, Title: title}, nil
}
