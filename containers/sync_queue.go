// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package containers

import (
	"sync"
	"time"
)

type SyncQueue struct {
	sync.Mutex
	Items []any
}

func (q *SyncQueue) Push(item any) {
	q.Lock()
	defer q.Unlock()
	q.Items = append(q.Items, item)
}

func (q *SyncQueue) Pop() any {
	q.Lock()
	defer q.Unlock()
	if len(q.Items) == 0 {
		return nil
	}
	item := q.Items[0]
	q.Items[0] = nil
	q.Items = q.Items[1:]
	return item
}

func (q *SyncQueue) PopAtIndex(index int) any {
	q.Lock()
	defer q.Unlock()
	if index < 0 || index >= len(q.Items) {
		return nil
	}
	result := q.Items[index]
	q.Items[index] = nil
	q.Items = append(q.Items[:index], q.Items[index+1:]...)
	return result
}

func (q *SyncQueue) Length() int {
	return len(q.Items)
}

type SyncUniqueQueue struct {
	SyncQueue
	SetWithTimes sync.Map
}

func (q *SyncUniqueQueue) Push(item any) {
	if _, stored := q.SetWithTimes.LoadOrStore(item, time.Now().UnixMilli()); !stored {
		q.SyncQueue.Push(item)
	}
}

func (q *SyncUniqueQueue) Pop() any {
	item := q.SyncQueue.Pop()
	q.SetWithTimes.Delete(item)
	return item
}

func (q *SyncUniqueQueue) PopAtIndex(index int) any {
	item := q.SyncQueue.PopAtIndex(index)
	if item != nil {
		q.SetWithTimes.Delete(item)
	}
	return item
}
