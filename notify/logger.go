package notify

import (
	"fmt"

	"github.com/kr/pretty"
	"github.com/fiatjaf/ilno/event"
	"github.com/fiatjaf/ilno/ilno"
	"github.com/fiatjaf/ilno/logger"
)

// Logger log notifications
type Logger struct{}

// Register Subscribe events
func (l *Logger) Register(eb *event.Bus) {
	eb.Subscribe("comments.new:new-thread", l.newThread)
	eb.Subscribe("comments.new:finish", l.newComment)
	eb.Subscribe("comments.edit", l.editComment)
	eb.Subscribe("comments.delete", l.deleteComment)
	eb.Subscribe("comments.activate", l.activateComment)
}

func (l *Logger) newThread(mt ilno.Thread) {
	logger.Info("new thread %s: %s", mt.ID, mt.Title)
}

func (l *Logger) newComment(c ilno.Comment) {
	logger.Info(fmt.Sprintf("create comment %# v", pretty.Formatter(c)))
}

func (l *Logger) editComment(id int) {
	logger.Info("comment edited %d: ", id)
}

func (l *Logger) deleteComment(id int) {
	logger.Info("comment deleted %d: ", id)
}

func (l *Logger) activateComment(id int) {
	logger.Info("comment %d activated: ", id)
}
