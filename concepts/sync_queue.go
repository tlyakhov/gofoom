package concepts

import "sync"

type SyncQueue struct {
	sync.Mutex
	Items []interface{}
}

func (q *SyncQueue) Push(item interface{}) {
	q.Lock()
	defer q.Unlock()
	q.Items = append(q.Items, item)
}

func (q *SyncQueue) Pop() interface{} {
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

func (q *SyncQueue) Length() int {
	return len(q.Items)
}

type SyncUniqueQueue struct {
	SyncQueue
	set sync.Map
}

func (q *SyncUniqueQueue) Push(item interface{}) {
	if _, stored := q.set.LoadOrStore(item, true); !stored {
		q.SyncQueue.Push(item)
	}
}

func (q *SyncUniqueQueue) Pop() interface{} {
	item := q.SyncQueue.Pop()
	q.set.Delete(item)
	return item
}
