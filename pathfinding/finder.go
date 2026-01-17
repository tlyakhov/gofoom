package pathfinding

import (
	"container/heap"
	"math"
	"slices"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
)

type Finder struct {
	Start       *concepts.Vector3
	End         *concepts.Vector3
	Step        float64
	Radius      float64
	StartSector *core.Sector
	SectorValid func(from *core.Sector, to *core.Sector, p *concepts.Vector2) bool
}

// Directions: 8 neighbors
var directions = []struct{ dx, dy int }{
	{1, 0}, {-1, 0}, {0, 1}, {0, -1},
	{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
}

func (f *Finder) sectorForNextPoint(sector *core.Sector, delta, next *concepts.Vector2, depth int) *core.Sector {
	if sector.IsPointInside2D(next) {
		for _, seg := range sector.Segments {
			if seg.AdjacentSegment != nil && seg.PortalIsPassable {
				continue
			}
			if seg.DistanceToPointSq(next) < f.Radius*f.Radius {
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
		if f.SectorValid != nil && !f.SectorValid(sector, adj, next) {
			continue
		}

		if adj = f.sectorForNextPoint(seg.AdjacentSegment.Sector, delta, next, depth+1); adj != nil {
			return adj
		}
	}
	return nil
}

// Helper to convert grid key to world point
func (f *Finder) keyToPoint(k nodeKey) concepts.Vector3 {
	return concepts.Vector3{
		f.Start[0] + float64(k.x)*f.Step,
		f.Start[1] + float64(k.y)*f.Step,
	}
}

// ShortestPath finds the shortest path between start and end points using the
// A* algorithm. It builds the graph on the fly by moving in fixed increments
// from the start point.
func (f *Finder) ShortestPath() []concepts.Vector3 {
	//defer concepts.ExecutionDuration(concepts.ExecutionTrack("PathFinder.ShortestPath"))
	if f.Step <= 0 {
		return nil
	}

	sector := f.StartSector
	if sector == nil {
		// TODO: Optimize these checks
		sector = findSectorForPoint(f.Start)
		// Check if start is valid
		if sector == nil {
			//log.Printf("PathFinder.ShortestPath: Invalid starting point: %v, %v", pf.Start, pf.End)
			return nil
		}
	}

	// cameFrom maps a node to its predecessor
	cameFrom := make(map[nodeKey]*node)
	// costSoFar stores the g cost
	costSoFar := make(map[nodeKey]float64)

	// TODO: Investigate memory usage here. With a lot of path finders and large
	// levels, we could have a lot of nodes
	// Open set
	pq := make(priorityQueue, 0)
	heap.Init(&pq)

	startKey := nodeKey{0, 0}
	startNode := &node{
		key:           startKey,
		sector:        sector,
		totalCost:     0,
		costFromStart: 0,
		costFromEnd:   f.Start.Dist(f.End),
	}
	heap.Push(&pq, startNode)
	costSoFar[startKey] = 0

	var finalKey *nodeKey
	path := []concepts.Vector3{}
	lowestCost := math.Inf(1)

	// Use a special key for the exact end point
	endKey := nodeKey{math.MaxInt, math.MaxInt}

	for pq.Len() > 0 {
		current := heap.Pop(&pq).(*node)
		currentPoint := f.keyToPoint(current.key)
		delta := concepts.Vector2{}
		// Check if we are close enough to end to jump there directly
		// Using 1.5 * stepSize to cover diagonals and a bit of slack
		if currentPoint.DistSq(f.End) <= (f.Step * 1.5 * f.Step * 1.5) {
			cameFrom[endKey] = current
			finalKey = &endKey
			path = append(path, *f.End)
			break
		}
		if current.costFromEnd < lowestCost {
			lowestCost = current.costFromEnd
			finalKey = &current.key
		}

		for _, d := range directions {
			nextKey := nodeKey{current.key.x + d.dx, current.key.y + d.dy}
			delta[0] = float64(d.dx)
			delta[1] = float64(d.dy)
			nextPoint := f.keyToPoint(nextKey)
			nextSector := f.sectorForNextPoint(current.sector, &delta, nextPoint.To2D(), 0)
			if nextSector == nil {
				continue
			}

			// Cost is strictly step size (or diagonal step size)
			dist := f.Step
			if d.dx != 0 && d.dy != 0 {
				dist = f.Step * math.Sqrt2
			}
			nextCostFromStart := current.costFromStart + dist

			if prevCost, exists := costSoFar[nextKey]; !exists || nextCostFromStart < prevCost {
				costSoFar[nextKey] = nextCostFromStart
				costFromEnd := nextPoint.Dist(f.End)
				totalCost := nextCostFromStart + costFromEnd // Heuristic: Euclidean distance
				heap.Push(&pq, &node{
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
		path = append(path, f.keyToPoint(key))
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

func findSectorForPoint(p *concepts.Vector3) *core.Sector {
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
