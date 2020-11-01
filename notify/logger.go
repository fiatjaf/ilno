package notify

import (
	"fmt"

	"github.com/fiatjaf/ilno/event"
	"github.com/fiatjaf/ilno/ilno"
	"github.com/fiatjaf/ilno/logger"
	"github.com/kr/pretty"
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
	eb.Subscribe("admin.ban", l.banUser)
	eb.Subscribe("admin.unban", l.unbanUser)
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

func (l *Logger) banUser(key string) {
	logger.Info("user %s banned", key)
}

func (l *Logger) unbanUser(key string) {
	logger.Info("user %s unbanned", key)
}
