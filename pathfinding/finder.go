// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

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
	MountHeight float64
	Ray         concepts.Ray
}

// Directions: 8 neighbors
var directions = []struct{ dx, dy int }{
	{1, 0}, {-1, 0}, {0, 1}, {0, -1},
	{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
}

func (f *Finder) sectorForNextPoint(startSector *core.Sector, start, next *concepts.Vector3) *core.Sector {
	dir := next.Sub(start)
	dist := dir.Length()
	if dist < constants.IntersectEpsilon {
		return startSector
	}
	dir.MulSelf(1.0 / dist)

	// Check radius by extending the ray beyond the destination
	limit := dist + f.Radius
	limitSq := limit * limit

	f.Ray.Start = *start
	f.Ray.Delta = *dir
	f.Ray.Limit = limit

	f.Ray.End = f.Ray.Start
	f.Ray.End.AddSelf(f.Ray.Delta.Mul(f.Ray.Limit))

	var req core.CastRequest
	req.Ray = &f.Ray
	req.MinDistSq = -1.0

	sector := startSector
	depth := 0

	for sector != nil {
		req.HitDistSq = limitSq
		req.HitSegment = nil
		req.NextSector = nil

		// 1. Check Exit (check boundaries of current sector)
		req.CheckEntry = false
		sector.IntersectRay(&req)

		// 2. Check Entry (check higher layers)
		for _, e := range sector.HigherLayers {
			if e == 0 {
				continue
			}
			overlap := core.GetSector(e)
			if overlap == nil {
				continue
			}
			req.CheckEntry = true
			overlap.IntersectRay(&req)
		}

		bestHit := req.CastResponse

		// If no hit found within limit, the path is clear.
		// We remain in the current sector (or the last traversed sector).
		if bestHit.HitSegment == nil {
			return sector
		}

		// We hit something.
		if bestHit.NextSector == nil {
			// Hit a solid wall. Path is blocked.
			return nil
		}

		// Hit a portal. Check validity.
		hitPoint2D := bestHit.HitPoint.To2D()
		if f.SectorValid != nil && !f.SectorValid(sector, bestHit.NextSector, hitPoint2D) {
			return nil
		}

		// Traverse to next sector
		sector = bestHit.NextSector
		req.MinDistSq = bestHit.HitDistSq
		depth++
		if depth > constants.MaxPortals {
			return nil
		}
	}

	return nil
}

// Helper to convert grid key to world point
func (f *Finder) keyToPoint(k nodeKey) concepts.Vector3 {
	return concepts.Vector3{
		f.Start[0] + float64(k.x)*f.Step,
		f.Start[1] + float64(k.y)*f.Step,
		f.Start[2],
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

	// If MountHeight is set, use it. Otherwise, use a small epsilon.
	zBias := 0.1
	if f.MountHeight > 0 {
		zBias = f.MountHeight
	}

	for pq.Len() > 0 {
		current := heap.Pop(&pq).(*node)
		currentPoint := f.keyToPoint(current.key)

		// Adjust Z based on current sector to ensure raycast works
		if current.sector != nil {
			fz, _ := current.sector.ZAt(currentPoint.To2D())
			// Bias up to avoid floor intersection issues
			currentPoint[2] = fz + zBias
		}

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
			nextPoint := f.keyToPoint(nextKey)
			// Assume horizontal step
			nextPoint[2] = currentPoint[2]

			nextSector := f.sectorForNextPoint(current.sector, &currentPoint, &nextPoint)
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
		// Use correct Z for path points too?
		// keyToPoint uses Start[2]. This might be off if we went up/down.
		// However, path reconstruction usually cares about X/Y.
		// If we want 3D path, we should probably store Z in the node or cameFrom map.
		// But nodeKey is 2D.
		// We'll stick to 2D approximation for output, or just use keyToPoint.
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
