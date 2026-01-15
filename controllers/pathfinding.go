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
	Start       concepts.Vector2
	End         concepts.Vector2
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

func (pf *PathFinder) checkPath(startSector *core.Sector, start, end *concepts.Vector2) (*core.Sector, bool) {
	currSector := startSector
	currPos := *start
	target := *end

	// Safety break
	for i := 0; i < 20; i++ {
		// Find closest intersection with sector boundaries
		var closestSeg *core.SectorSegment
		closestDist := math.MaxFloat64

		// Ray cast against all segments
		rayDelta := target.Sub(&currPos)
		distTotal := rayDelta.Length()

		if distTotal < constants.IntersectEpsilon {
			// Already there
			if currSector.IsPointInside2D(&target) {
				return currSector, true
			}
			return nil, false
		}

		for _, seg := range currSector.Segments {
			// We use IntersectSegmentsRaw directly to get the ray parameter 's'
			// seg.A, seg.B are the segment points.
			// currPos, target are the ray points.
			// The function returns r (segment param) and s (ray param).
			// If r is in [0, 1] and s is in [0, 1], we have an intersection.
			// However, IntersectSegmentsRaw might return -1 if no intersection.
			// Wait, IntersectSegmentsRaw returns 4 values.
			// r, s, dx, dy.
			// If no intersection, it returns -1.

			// We need to re-implement raw call because concepts pkg might clamp or reject if not exact.
			// Actually IntersectSegmentsRaw logic:
			// if r or s out of range, returns -1.
			// We want intersections strictly on the segment (r in [0, 1]).
			// And on the ray (s in [0, 1]).
			// This matches exactly what we need.

			r, s, _, _ := concepts.IntersectSegmentsRaw(seg.A, seg.B, &currPos, &target)

			if r >= 0 {
				// Intersection found.
				// s is the fraction along the ray (currPos -> target).
				// 0 <= s <= 1.
				dist := s * distTotal
				if dist < closestDist {
					closestDist = dist
					closestSeg = seg
				}
			}
		}

		if closestSeg == nil {
			// No walls hit.
			// Check if target is inside.
			if currSector.IsPointInside2D(&target) {
				// Valid!
				// One final check: DistanceToPointSq for walls at destination?
				for _, seg := range currSector.Segments {
					if seg.DistanceToPointSq(&target) < pf.Radius*pf.Radius {
						// If it's a portal, maybe ok?
						if seg.AdjacentSegment != nil && seg.PortalIsPassable {
							continue
						}
						return nil, false
					}
				}
				return currSector, true
			}
			// Not inside, but no intersection?
			// This can happen if start was outside or due to precision.
			// Assuming robust geometry, return false.
			return nil, false
		}

		// We hit a segment.
		// If it is very close to target (s ~ 1), we are good?
		// But IntersectSegmentsRaw guarantees s <= 1.
		// If s is very close to 1, we hit the wall AT the target.
		// Which is a collision.

		if closestDist >= distTotal-constants.IntersectEpsilon {
			// Hit is basically at target.
			// Collision.
			// Unless it is a portal we can pass through?
			// But target is exactly on the portal/wall.
			// Let's assume invalid if we end UP on a wall.
			return nil, false
		}

		// We hit a wall/portal BEFORE target.
		if closestSeg.AdjacentSegment != nil && closestSeg.PortalIsPassable {
			// It is a portal.
			nextSector := closestSeg.AdjacentSegment.Sector

			// Calculate intersection point exactly
			// intersection = currPos + (target - currPos) * (closestDist / distTotal)
			// Or just use 's' if we had it.
			// Recompute s for closestSeg?
			// s = closestDist / distTotal.
			s := closestDist / distTotal
			intersection := *currPos.Add(rayDelta.Mul(s))

			// Check Height / Door Logic AT THE PORTAL
			fz, cz := currSector.ZAt(&intersection)
			afz, acz := nextSector.ZAt(&intersection)

			if afz-fz > pf.MountHeight || min(cz, acz)-max(fz, afz) < pf.Radius {
				// Check for doors
				door := behaviors.GetDoor(currSector.Entity)
				adjDoor := behaviors.GetDoor(nextSector.Entity)
				if door != nil || adjDoor != nil {
					// TODO: Check for doors that NPCs can't walk through.
				} else {
					return nil, false
				}
			}

			// Traverse
			currSector = nextSector
			// Advance currPos slightly past portal to avoid hitting it again
			// Or just update currPos to intersection and trust the next iteration won't find the same wall behind us?
			// To be safe, nudge it.
			nudge := rayDelta.Norm().Mul(constants.IntersectEpsilon * 2.0)
			currPos = *intersection.Add(nudge)
			continue
		} else {
			// Hit a solid wall.
			return nil, false
		}
	}

	return nil, false
}

// Helper to convert grid key to world point
func (pf *PathFinder) keyToPoint(k pathNodeKey) concepts.Vector2 {
	return concepts.Vector2{
		pf.Start[0] + float64(k.x)*pf.Step,
		pf.Start[1] + float64(k.y)*pf.Step,
	}
}

// ShortestPath finds the shortest path between start and end points using the
// A* algorithm. It builds the graph on the fly by moving in fixed increments
// from the start point.
func (pf *PathFinder) ShortestPath() []concepts.Vector2 {
	defer concepts.ExecutionDuration(concepts.ExecutionTrack("PathFinder.ShortestPath"))
	if pf.Step <= 0 {
		return nil
	}

	// TODO: Optimize these checks
	sector := pathPointValid(&pf.Start)
	// Check if start and end are valid
	if sector == nil || pathPointValid(&pf.End) == nil {
		log.Printf("Invalid points: %v, %v", pf.Start, pf.End)
		return nil
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

	// Multipliers for dynamic step size
	multipliers := []int{4, 2, 1}

	for pq.Len() > 0 {
		current := heap.Pop(&pq).(*pathNode)
		currentPoint := pf.keyToPoint(current.key)
		// Check if we are close enough to end to jump there directly
		// Using 1.5 * stepSize to cover diagonals and a bit of slack
		if currentPoint.Dist(&pf.End) <= pf.Step*1.5 {
			cameFrom[endKey] = current.key
			finalKey = &endKey
			break
		}

		for _, d := range directions {
			// Try multipliers in descending order
			for _, mult := range multipliers {
				nextKey := pathNodeKey{current.key.x + d.dx*mult, current.key.y + d.dy*mult}
				// Skip if we already found a cheaper path to this exact node
				// Note: with multipliers, we might reach the same node via different steps.
				// e.g. 0->4 vs 0->2->4.

				// Wait, if we use multipliers, 'nextKey' is in grid coordinates (x/Step).
				// So if mult=4, we jump 4 units in grid.
				// We need to calculate cost properly.

				nextPoint := pf.keyToPoint(nextKey)
				nextSector, valid := pf.checkPath(current.sector, &currentPoint, &nextPoint)
				if !valid {
					continue
				}

				// Cost calculation
				dist := pf.Step * float64(mult)
				if d.dx != 0 && d.dy != 0 {
					dist *= math.Sqrt2
				}
				nextCostFromStart := current.costFromStart + dist

				if prevCost, exists := costSoFar[nextKey]; !exists || nextCostFromStart < prevCost {
					costSoFar[nextKey] = nextCostFromStart
					totalCost := nextCostFromStart + nextPoint.Dist(&pf.End) // Heuristic: Euclidean distance
					heap.Push(&pq, &pathNode{
						key:           nextKey,
						sector:        nextSector,
						totalCost:     totalCost,
						costFromStart: nextCostFromStart,
					})
					cameFrom[nextKey] = current.key
				}

				// If we successfully found a valid path with a large step,
				// we assume it is "good enough" for this direction and don't try smaller steps.
				// This optimizes performance by reducing branching factor.
				break
			}
		}
	}

	if finalKey == nil {
		log.Printf("finalKey = nil")
		return nil
	}

	// Reconstruct path
	path := []concepts.Vector2{pf.End}
	key := *finalKey
	ok := true

	// If we finished at the special end key, step back to the grid
	if key == endKey {
		key = cameFrom[key]
	}

	for {
		path = append(path, pf.keyToPoint(key))
		if key == startKey {
			break
		}
		key, ok = cameFrom[key]
		if !ok {
			break // Should not happen
		}
	}

	slices.Reverse(path)

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
