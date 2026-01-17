package controllers

import (
	"container/heap"
	"log"
	"math"
	"slices"

	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
)

type PathFinder struct {
	Start       *concepts.Vector3
	End         *concepts.Vector3
	Step        float64
	Radius      float64
	MountHeight float64
	StartSector *core.Sector
}

// pathNodeKey represents a coordinate on the virtual grid.
type pathNodeKey struct {
	x, y int
}

// pathNode holds the state for the priority queue.
type pathNode struct {
	key           pathNodeKey
	sector        *core.Sector
	totalCost     float64 // f = g + h
	costFromStart float64
	costFromEnd   float64
	index         int // The index of the item in the heap
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

func (pf *PathFinder) sectorForNextPoint(sector *core.Sector, delta, next *concepts.Vector2, depth int) *core.Sector {
	if sector.IsPointInside2D(next) {
		for _, seg := range sector.Segments {
			if seg.AdjacentSegment != nil && seg.PortalIsPassable {
				continue
			}
			if seg.DistanceToPointSq(next) < pf.Radius*pf.Radius {
				return nil
			}
		}
		return sector
	}

	if depth > 2 {
		return nil
	}

	// TODO: Ignore dead ends (e.g. if there's only one portal segment
	// and the end point isn't in the adj sector, break early)

	fz, cz := sector.ZAt(next)
	for _, seg := range sector.Segments {
		if seg.AdjacentSegment == nil {
			continue
		}
		// Don't move backwards
		if seg.Normal[0]*delta[0]+seg.Normal[1]*delta[1] > 0 {
			continue
		}
		if seg.DistanceToPointSq(next) < constants.IntersectEpsilon {
			return sector
		}
		adj := seg.AdjacentSegment.Sector
		afz, acz := adj.ZAt(next)
		// Check that the sector height is mountable and isn't too narrow
		if afz-fz > pf.MountHeight || min(cz, acz)-max(fz, afz) < pf.Radius {
			// Maybe this or previous sector is a door?
			door := behaviors.GetDoor(sector.Entity)
			adjDoor := behaviors.GetDoor(adj.Entity)
			if door != nil || adjDoor != nil {
				// TODO: Check for doors that NPCs can't walk through.
			} else {
				// Not a door, impassable
				continue
			}
		}
		if adj = pf.sectorForNextPoint(seg.AdjacentSegment.Sector, delta, next, depth+1); adj != nil {
			return adj
		}
	}
	return nil
}

// Helper to convert grid key to world point
func (pf *PathFinder) keyToPoint(k pathNodeKey) concepts.Vector3 {
	return concepts.Vector3{
		pf.Start[0] + float64(k.x)*pf.Step,
		pf.Start[1] + float64(k.y)*pf.Step,
	}
}

// ShortestPath finds the shortest path between start and end points using the
// A* algorithm. It builds the graph on the fly by moving in fixed increments
// from the start point.
func (pf *PathFinder) ShortestPath() []concepts.Vector3 {
	//defer concepts.ExecutionDuration(concepts.ExecutionTrack("PathFinder.ShortestPath"))
	if pf.Step <= 0 {
		return nil
	}

	// TODO: Optimize these checks
	sector := pathPointValid(pf.Start)
	// Check if start is valid
	if sector == nil {
		log.Printf("Invalid point: %v, %v", pf.Start, pf.End)
		return nil
	}

	// cameFrom maps a node to its predecessor
	cameFrom := make(map[pathNodeKey]*pathNode)
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
		costFromEnd:   pf.Start.Dist(pf.End),
	}
	heap.Push(&pq, startNode)
	costSoFar[startKey] = 0

	var finalKey *pathNodeKey
	path := []concepts.Vector3{}
	lowestCost := math.Inf(1)

	// Use a special key for the exact end point
	endKey := pathNodeKey{math.MaxInt, math.MaxInt}

	for pq.Len() > 0 {
		current := heap.Pop(&pq).(*pathNode)
		currentPoint := pf.keyToPoint(current.key)
		delta := concepts.Vector2{}
		// Check if we are close enough to end to jump there directly
		// Using 1.5 * stepSize to cover diagonals and a bit of slack
		if currentPoint.DistSq(pf.End) <= (pf.Step * 1.5 * pf.Step * 1.5) {
			cameFrom[endKey] = current
			finalKey = &endKey
			path = append(path, *pf.End)
			break
		}
		if current.costFromEnd < lowestCost {
			lowestCost = current.costFromEnd
			finalKey = &current.key
		}

		for _, d := range directions {
			nextKey := pathNodeKey{current.key.x + d.dx, current.key.y + d.dy}
			delta[0] = float64(d.dx)
			delta[1] = float64(d.dy)
			nextPoint := pf.keyToPoint(nextKey)
			nextSector := pf.sectorForNextPoint(current.sector, &delta, nextPoint.To2D(), 0)
			if nextSector == nil {
				continue
			}

			// Cost is strictly step size (or diagonal step size)
			dist := pf.Step
			if d.dx != 0 && d.dy != 0 {
				dist = pf.Step * math.Sqrt2
			}
			nextCostFromStart := current.costFromStart + dist

			if prevCost, exists := costSoFar[nextKey]; !exists || nextCostFromStart < prevCost {
				costSoFar[nextKey] = nextCostFromStart
				costFromEnd := nextPoint.Dist(pf.End)
				totalCost := nextCostFromStart + costFromEnd // Heuristic: Euclidean distance
				heap.Push(&pq, &pathNode{
					key:           nextKey,
					sector:        nextSector,
					totalCost:     totalCost,
					costFromStart: nextCostFromStart,
					costFromEnd:   costFromEnd,
				})
				cameFrom[nextKey] = current
			}
		}
	}

	if finalKey == nil {
		return nil
	}

	// Reconstruct path
	key := *finalKey

	// If we finished at the special end key, step back to the grid
	if key == endKey {
		key = cameFrom[key].key
	}

	for {
		path = append(path, pf.keyToPoint(key))
		if key == startKey {
			break
		}
		node, ok := cameFrom[key]
		if !ok {
			break // Should not happen
		}
		key = node.key
	}

	slices.Reverse(path)

	return path
}

func pathPointValid(p *concepts.Vector3) *core.Sector {
	p2d := p.To2D()
	arena := ecs.ArenaFor[core.Sector](core.SectorCID)
	for i := range arena.Cap() {
		sector := arena.Value(i)
		if sector == nil {
			continue
		}
		if sector.IsPointInside2D(p2d) {
			return sector
		}
		for _, seg := range sector.Segments {
			if seg.AdjacentSector != 0 && seg.DistanceToPointSq(p2d) < constants.IntersectEpsilon {
				return sector
			}
		}
	}
	return nil
}
