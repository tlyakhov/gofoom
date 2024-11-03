// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"tlyakhov/gofoom/ecs"
)

// This stores a linked-list queue of sectors to recalculate,
// and also a map to remove duplicates
type PvsQueue struct {
	ecs.Attached `editable:"^"`

	Head    *pvsQueueNode
	Tail    *pvsQueueNode
	Sectors map[*Sector]*pvsQueueNode
}

type pvsQueueNode struct {
	*Sector
	Prev *pvsQueueNode
	Next *pvsQueueNode
}

var PvsQueueCID ecs.ComponentID

func init() {
	PvsQueueCID = ecs.RegisterComponent(&ecs.Column[PvsQueue, *PvsQueue]{Getter: GetPvsQueue})
}

func GetPvsQueue(db *ecs.ECS, e ecs.Entity) *PvsQueue {
	if asserted, ok := db.Component(e, PvsQueueCID).(*PvsQueue); ok {
		return asserted
	}
	return nil
}

func (q *PvsQueue) String() string {
	return "PvsQueue"
}

func (q *PvsQueue) PushHead(s *Sector) {
	if _, contains := q.Sectors[s]; contains {
		return
	}

	node := &pvsQueueNode{Sector: s, Prev: nil, Next: q.Head}
	q.Head = node
	q.Sectors[s] = node
	if node.Next != nil {
		node.Next.Prev = node
	} else {
		q.Tail = node
	}
}

func (q *PvsQueue) PushTail(s *Sector) {
	if _, contains := q.Sectors[s]; contains {
		return
	}

	node := &pvsQueueNode{Sector: s, Prev: q.Tail, Next: nil}
	q.Tail = node
	q.Sectors[s] = node
	if node.Prev != nil {
		node.Prev.Next = node
	} else {
		q.Head = node
	}
}

func (q *PvsQueue) Contains(s *Sector) bool {
	_, contains := q.Sectors[s]
	return contains
}

func (q *PvsQueue) PopHead() *Sector {
	head := q.Head
	if head == nil {
		return nil
	}
	q.Head = head.Next
	s := head.Sector
	delete(q.Sectors, s)

	if head.Next != nil {
		head.Next.Prev = nil
	} else {
		q.Tail = nil
	}
	head.Sector = nil
	head.Next = nil
	return s
}

func (q *PvsQueue) PopTail() *Sector {
	tail := q.Tail
	if tail == nil {
		return nil
	}
	q.Tail = tail.Prev
	s := tail.Sector
	delete(q.Sectors, s)

	if tail.Prev != nil {
		tail.Prev.Next = nil
	} else {
		q.Head = nil
	}
	tail.Sector = nil
	tail.Prev = nil
	return s
}

func (q *PvsQueue) Construct(data map[string]any) {
	q.Attached.Construct(data)

	q.Sectors = make(map[*Sector]*pvsQueueNode)
}

func (q *PvsQueue) Serialize() map[string]any {
	result := q.Attached.Serialize()

	return result
}
