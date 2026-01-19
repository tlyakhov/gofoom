// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package pathfinding

type priorityQueue []*node

func (q priorityQueue) Len() int { return len(q) }

func (q priorityQueue) Less(i, j int) bool {
	return q[i].totalCost < q[j].totalCost
}

func (q priorityQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q *priorityQueue) Push(x any) {
	n := len(*q)
	item := x.(*node)
	item.index = n
	*q = append(*q, item)
}

func (q *priorityQueue) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*q = old[0 : n-1]
	return item
}
