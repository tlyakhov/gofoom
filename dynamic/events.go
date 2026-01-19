// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package dynamic

import (
	"log"
	"sync"
)

type EventID int

type Event struct {
	ID           EventID
	Timestamp    int64 // Milliseconds
	SimTimestamp int64 // Milliseconds
	Data         any
}

type EventClass struct {
	ID        EventID
	Name      string
	Consumers []EventConsumer
}

const MaxEvents = 1024

type EventConsumer func(evt *Event) bool
type EventQueue struct {
	Queue [MaxEvents]*Event
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

func SubscribeToEvent(id EventID, c EventConsumer) {
	eventClassLock.Lock()
	defer eventClassLock.Unlock()
	eventClasses[id].Consumers = append(eventClasses[id].Consumers, c)
}

// TODO: Unsubscribe

func (q *EventQueue) PushEvent(evt *Event) {
	q.Queue[q.Tail] = evt
	q.Tail = (q.Tail + 1) % MaxEvents
	if q.Tail == q.Head {
		log.Printf("Warning: too many events for queue size %v.", MaxEvents)
		q.Head = (q.Head + 1) % MaxEvents
	}
}

func (q *EventQueue) popEvent() *Event {
	if q.Head == q.Tail {
		return nil
	}
	evt := q.Queue[q.Head]
	q.Head = (q.Head + 1) % MaxEvents
	return evt
}

func (q *EventQueue) ConsumeEvent() {
	if q.Head == q.Tail {
		return
	}
	evt := q.popEvent()
	for _, c := range eventClasses[evt.ID].Consumers {
		if c(evt) {
			return
		}
	}
}
