// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package containers

import (
	"sync"
	"time"
)

type SyncQueue[T comparable] struct {
	sync.Mutex
	Items []T
}

func (q *SyncQueue[T]) Push(item T) {
	q.Lock()
	defer q.Unlock()
	q.Items = append(q.Items, item)
}

func (q *SyncQueue[T]) PushHead(item T) {
	q.Lock()
	defer q.Unlock()
	var empty T
	q.Items = append(q.Items, empty)
	copy(q.Items[1:], q.Items)
	q.Items[0] = item
}

func (q *SyncQueue[T]) Pop() T {
	q.Lock()
	defer q.Unlock()
	if len(q.Items) == 0 {
		var empty T
		return empty
	}
	item := q.Items[0]
	var empty T
	q.Items[0] = empty
	q.Items = q.Items[1:]
	return item
}

func (q *SyncQueue[T]) PopAtIndex(index int) T {
	q.Lock()
	defer q.Unlock()
	if index < 0 || index >= len(q.Items) {
		var empty T
		return empty
	}
	result := q.Items[index]
	var empty T
	q.Items[index] = empty
	q.Items = append(q.Items[:index], q.Items[index+1:]...)
	return result
}

func (q *SyncQueue[T]) Length() int {
	return len(q.Items)
}

type SyncUniqueQueue[T comparable] struct {
	SyncQueue[T]
	SetWithTimes sync.Map
}

func (q *SyncUniqueQueue[T]) Push(item T) {
	if _, stored := q.SetWithTimes.LoadOrStore(item, time.Now().UnixMilli()); !stored {
		q.SyncQueue.Push(item)
	}
}

func (q *SyncUniqueQueue[T]) Pop() any {
	item := q.SyncQueue.Pop()
	q.SetWithTimes.Delete(item)
	return item
}

func (q *SyncUniqueQueue[T]) PopAtIndex(index int) T {
	item := q.SyncQueue.PopAtIndex(index)
	var empty T
	if item != empty {
		q.SetWithTimes.Delete(item)
	}
	return item
}
