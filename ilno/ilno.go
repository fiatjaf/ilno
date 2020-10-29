package ilno

import (
	"context"

	"github.com/fiatjaf/ilno/config"
	"github.com/fiatjaf/ilno/event"
	"github.com/fiatjaf/ilno/tool/markdown"
)

const descStorageNotFound = "no result found in storage"
const descStorageUnhandledError = "storage raise unhandled error"
const descRequestInvalidParm = "can not parse parameters correctly"
const descRequestInvalidCookies = "invalid cookies in request(or missing cookies)"

type ilnoContextKey int

// ILNOContextKey can be used as key for context
var ILNOContextKey ilnoContextKey = 1

// RequestIDFromContext return request id from Context
func RequestIDFromContext(ctx context.Context) string {
	requestID, ok := ctx.Value(ILNOContextKey).(string)
	if !ok {
		requestID = "unknown"
	}
	return requestID
}

// ILNO do the main logical staff
type ILNO struct {
	storage Storage
	config  config.Config
	tools   tools
}

type tools struct {
	markdown *markdown.Worker
	event    *event.Bus
}

// New a ILNO instance
func New(cfg config.Config, storage Storage) *ILNO {
	return &ILNO{
		config: cfg,
		tools: tools{
			markdown: markdown.New(),
			event:    event.New(),
		},
		storage: storage,
	}
}
