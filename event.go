package ekanite

import (
	"fmt"

	"github.com/ekanite/ekanite/input"
)

// Event is a log message that can be indexed.
type Event struct {
	*input.Event
}

// NewEvent returns a new Event.
func NewEvent() *Event {
	return &Event{}
}

// ID returns a unique ID for the event.
func (e Event) ID() DocID {
	return DocID(fmt.Sprintf("%016x%016x",
		uint64(e.ReferenceTime().UnixNano()), uint64(e.Sequence)))
}

// Data returns the indexable data.
func (e Event) Data() interface{} {
	return struct {
		Message       string
		ReferenceTime string
	}{
		Message:       e.Text,
		ReferenceTime: e.ReferenceTime().Format("2006-01-02T15:04:05"),
	}
}

// Source returns the original received data.
func (e Event) Source() []byte {
	return []byte(e.Text)
}
