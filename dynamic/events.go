package dynamic

import "sync"

type EventID int

type Event struct {
	ID        EventID
	Timestamp int64 // Milliseconds
	Params    []any
}

type EventClass struct {
	ID         EventID
	Name       string
	ParamCount int
}

const MaxEvents = 1024

type EventQueue struct {
	Queue [MaxEvents]Event
	Head  int
	Tail  int
}

var eventClassLock sync.Mutex
var eventClasses = []*EventClass{nil}

func RegisterEventClass(ec *EventClass) EventID {
	eventClassLock.Lock()
	defer eventClassLock.Unlock()
	id := EventID(len(eventClasses))
	ec.ID = id
	eventClasses = append(eventClasses, ec)
	return id
}

func EventClasses() []*EventClass {
	return eventClasses
}
