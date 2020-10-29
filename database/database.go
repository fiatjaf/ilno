package database

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"path"
	"runtime"
	"strings"
	"time"

	// sqlite3 driver
	"github.com/fiatjaf/ilno/ilno"
	"github.com/fiatjaf/ilno/logger"
	"github.com/fiatjaf/ilno/tool/bloomfilter"
	"github.com/fiatjaf/ilno/version"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/guregu/null.v4"
)

// Database handles all operations related to the database.
type Database struct {
	*sql.DB
	statement map[string]string
	timeout   time.Duration
}

type databaseError struct {
	caller string
	file   string
	line   int
	origin error
}

func (de databaseError) Error() string {
	return fmt.Sprintf("%s: %v", de.caller, de.origin)
}

// Format formats the error according to the fmt.Formatter interface.
func (de databaseError) Format(s fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		switch {
		case s.Flag('+'):
			io.WriteString(s, fmt.Sprintf("'%s:%d %s' %v", path.Base(de.file), de.line, de.caller, de.origin))
		default:
			io.WriteString(s, de.Error())
		}
	}
}

func (de databaseError) Unwrap() error {
	return de.origin
}

func wraperror(err error) databaseError {
	if err == sql.ErrNoRows {
		err = ilno.ErrStorageNotFound
	}

	var caller string
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		caller = "unkown"
	} else {
		fn := runtime.FuncForPC(pc)
		caller = fn.Name()
	}

	return databaseError{
		origin: err,
		caller: strings.TrimPrefix(caller, version.Mod),
		file:   file,
		line:   line,
	}
}

// New return a *Database
func New(path string, timeout time.Duration) (*Database, error) {
	databaseType := "sqlite3"
	if path == "" {
		path = ":memory:"
	}

	db, err := sql.Open(databaseType, path)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	_, err = db.Exec(presetSQL[databaseType]["create"])
	if err != nil {
		return nil, err
	}
	logger.Debug("create database instance at %s", path)
	return &Database{db, presetSQL[databaseType], timeout}, nil
}

func (d *Database) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, d.timeout)
}

type nullComment struct {
	TID          int64
	ID           int64
	Parent       null.Int
	Created      float64
	Modified     null.Float
	Mode         int
	Text         string
	Key          string
	Author       string
	Likes        int
	Dislikes     int
	Voters       []byte
	Notification int
}

func (nc nullComment) ToComment() ilno.Comment {
	c := ilno.Comment{
		ID:       nc.ID,
		Parent:   &nc.Parent.Int64,
		Created:  nc.Created,
		Modified: &nc.Modified.Float64,
		Mode:     nc.Mode,
		Text:     nc.Text,
		Key:      nc.Key,
		Author:   nc.Author,
		Likes:    nc.Likes,
		Dislikes: nc.Dislikes,
	}
	copy(c.Voters[:], nc.Voters)
	if !nc.Parent.Valid {
		c.Parent = nil
	}
	if !nc.Modified.Valid {
		c.Modified = nil
	}
	return c
}

func newNullComment(c ilno.Comment, threadID int64) nullComment {
	bf := bloomfilter.New()

	v := bf.Buffer()
	voters := make([]byte, 256)
	copy(voters, v[:])
	return nullComment{
		TID:      threadID,
		ID:       c.ID,
		Parent:   null.IntFromPtr(c.Parent),
		Created:  float64(time.Now().UnixNano()) / float64(1e9),
		Modified: null.NewFloat(0, false),
		Mode:     c.Mode,
		Text:     c.Text,
		Key:      c.Key,
		Author:   c.Author,
		Likes:    c.Likes,
		Dislikes: c.Dislikes,
		Voters:   voters,
	}
}

func (d *Database) execstmt(ctx context.Context, rowsaffected *int64, lastinsertid *int64, stmt string, args ...interface{}) error {
	result, err := d.DB.ExecContext(ctx, stmt, args...)
	if err != nil {
		return err
	}
	if rowsaffected != nil {
		*rowsaffected, err = result.RowsAffected()
		if err != nil {
			return err
		}
	}
	if lastinsertid != nil {
		*lastinsertid, err = result.LastInsertId()
		if err != nil {
			return err
		}
	}
	return nil
}
