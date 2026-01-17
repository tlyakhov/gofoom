package pathfinding

import "tlyakhov/gofoom/components/core"

// nodeKey represents a coordinate on the virtual grid.
type nodeKey struct {
	x, y int
}

// node holds the state for the priority queue.
type node struct {
	key           nodeKey
	sector        *core.Sector
	totalCost     float64 // f = g + h
	costFromStart float64
	costFromEnd   float64
	index         int // The index of the item in the heap
}
