package controllers

import (
	"container/heap"
	"log"
	"math"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
)

// pathNodeKey represents a coordinate on the virtual grid.
type pathNodeKey struct {
	x, y int
}

// pathNode holds the state for the priority queue.
type pathNode struct {
	key           pathNodeKey
	sector        *core.Sector
	totalCost     float64 // f = g + h
	costFromStart float64 // cost from start
	index         int     // The index of the item in the heap
}

type pathQueue []*pathNode

func (pq pathQueue) Len() int { return len(pq) }

func (pq pathQueue) Less(i, j int) bool {
	return pq[i].totalCost < pq[j].totalCost
}

func (pq pathQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *pathQueue) Push(x any) {
	n := len(*pq)
	item := x.(*pathNode)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *pathQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// Directions: 8 neighbors
var directions = []struct{ dx, dy int }{
	{1, 0}, {-1, 0}, {0, 1}, {0, -1},
	{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
}

func nextPointValid(sector *core.Sector, next *concepts.Vector2, depth int) *core.Sector {
	if sector.IsPointInside2D(next) {
		return sector
	}

	if depth > 2 {
		return nil
	}

	for _, seg := range sector.Segments {
		if seg.AdjacentSegment == nil {
			continue
		}
		if seg.DistanceToPointSq(next) < constants.IntersectEpsilon {
			return sector
		}
		if adj := nextPointValid(seg.AdjacentSegment.Sector, next, depth+1); adj != nil {
			return adj
		}
	}
	return nil
}

// ShortestPath finds the shortest path between start and end points using the
// A* algorithm. It builds the graph on the fly by moving in fixed increments
// from the start point.
func ShortestPath(start, end concepts.Vector2, stepSize float64) []concepts.Vector2 {
	if stepSize <= 0 {
		return nil
	}

	sector := pathPointValid(&start)
	// Check if start and end are valid
	if sector == nil || pathPointValid(&end) == nil {
		log.Printf("Invalid points: %v, %v", start, end)
		return nil
	}

	// Helper to convert grid key to world point
	keyToPoint := func(k pathNodeKey) concepts.Vector2 {
		return concepts.Vector2{
			start[0] + float64(k.x)*stepSize,
			start[1] + float64(k.y)*stepSize,
		}
	}

	// cameFrom maps a node to its predecessor
	cameFrom := make(map[pathNodeKey]pathNodeKey)
	// costSoFar stores the g cost
	costSoFar := make(map[pathNodeKey]float64)

	// Open set
	pq := make(pathQueue, 0)
	heap.Init(&pq)

	startKey := pathNodeKey{0, 0}
	startNode := &pathNode{
		key:           startKey,
		sector:        sector,
		totalCost:     0,
		costFromStart: 0,
	}
	heap.Push(&pq, startNode)
	costSoFar[startKey] = 0

	var finalKey *pathNodeKey
	// Use a special key for the exact end point
	endKey := pathNodeKey{math.MaxInt, math.MaxInt}

	for pq.Len() > 0 {
		current := heap.Pop(&pq).(*pathNode)
		currentPoint := keyToPoint(current.key)

		// Check if we are close enough to end to jump there directly
		// Using 1.5 * stepSize to cover diagonals and a bit of slack
		if currentPoint.Dist(&end) <= stepSize*1.5 {
			cameFrom[endKey] = current.key
			finalKey = &endKey
			break
		}

		for _, d := range directions {
			nextKey := pathNodeKey{current.key.x + d.dx, current.key.y + d.dy}
			nextPoint := keyToPoint(nextKey)

			nextSector := nextPointValid(current.sector, &nextPoint, 0)
			if nextSector == nil {
				continue
			}

			// Cost is strictly step size (or diagonal step size)
			dist := stepSize
			if d.dx != 0 && d.dy != 0 {
				dist = stepSize * math.Sqrt2
			}
			nextCostFromStart := current.costFromStart + dist

			if prevCost, exists := costSoFar[nextKey]; !exists || nextCostFromStart < prevCost {
				costSoFar[nextKey] = nextCostFromStart
				totalCost := nextCostFromStart + nextPoint.Dist(&end) // Heuristic: Euclidean distance
				heap.Push(&pq, &pathNode{
					key:           nextKey,
					sector:        nextSector,
					totalCost:     totalCost,
					costFromStart: nextCostFromStart,
				})
				cameFrom[nextKey] = current.key
			}
		}
	}

	if finalKey == nil {
		log.Printf("finalKey = nil")
		return nil
	}

	// Reconstruct path
	path := []concepts.Vector2{end}
	key := *finalKey
	ok := true

	// If we finished at the special end key, step back to the grid
	if key == endKey {
		key = cameFrom[key]
	}

	for {
		path = append(path, keyToPoint(key))
		if key == startKey {
			break
		}
		key, ok = cameFrom[key]
		if !ok {
			break // Should not happen
		}
	}

	// Reverse path
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}

func pathPointValid(p *concepts.Vector2) *core.Sector {
	arena := ecs.ArenaFor[core.Sector](core.SectorCID)
	for i := range arena.Cap() {
		sector := arena.Value(i)
		if sector == nil {
			continue
		}
		if sector.IsPointInside2D(p) {
			return sector
		}
		for _, seg := range sector.Segments {
			if seg.AdjacentSector != 0 && seg.DistanceToPointSq(p) < constants.IntersectEpsilon {
				return sector
			}
		}
	}
	return nil
}
