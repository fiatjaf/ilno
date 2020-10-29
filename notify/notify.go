package notify

import "github.com/fiatjaf/ilno/event"

// Notifier register handlers to *event.Bus
type Notifier interface {
	Register(*event.Bus)
}
