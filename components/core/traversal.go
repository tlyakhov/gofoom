// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type RayIntersection struct {
	// Inputs
	concepts.Ray

	IgnoreSegment *Segment
	MinDistSq     float64
	MaxDistSq     float64
	CheckEntry    bool

	// Outputs
	HitSegment *SectorSegment
	HitPoint   concepts.Vector3
	HitDistSq  float64
	NextSector *Sector
}

// IntersectRay finds the nearest intersection of a ray with the sector's segments.
// It handles portal transitions, higher layer checks, and grazing ray edge cases.
//
// Arguments:
//   - ri: A RayIntersection struct containing ray definition, traversal state, and output fields.
//     The function will update ri.HitSegment, ri.HitPoint, ri.HitDistSq, and ri.NextSector.
func (s *Sector) IntersectRay(ri *RayIntersection) {
	ri.HitSegment = nil
	ri.NextSector = nil
	ri.HitDistSq = ri.MaxDistSq

	// Pre-declare variables for loop
	var intersectionTest concepts.Vector3
	var adj *Sector
	var floorZ, ceilZ, floorZ2, ceilZ2, intersectionDistSq float64

	for _, seg := range s.Segments {
		// Don't intersect with the segment we started on
		if &seg.Segment == ri.IgnoreSegment {
			continue
		}

		// When checking higher layers (Entry check), we ignore portals in that layer.
		// We only care about entering the sector via a solid boundary (conceptually).
		if ri.CheckEntry && seg.AdjacentSector != 0 {
			continue
		}

		// Check normal facing.
		// Ray vs Normal dot product.
		// Ray is rayDelta. Segment normal is seg.Normal.
		// We want:
		// Exit (checkEntry=false): Ray going OUT. Normal points IN. Angle obtuse. Dot < 0.
		// Entry (checkEntry=true): Ray going IN. Normal points IN. Angle acute. Dot > 0.
		//
		// logic: normalFacing := dot > 0
		// if checkEntry != normalFacing => continue
		// Entry (true) != Dot>0 (true) => false (keep).
		// Exit (false) != Dot>0 (false) => false (keep).
		dot := ri.Delta[0]*seg.Normal[0] + ri.Delta[1]*seg.Normal[1]
		normalFacing := dot > 0

		if ri.CheckEntry != normalFacing {
			continue
		}

		// Find the intersection with this segment.
		if !seg.Intersect3D(&ri.Start, &ri.End, &intersectionTest) {
			continue
		}

		// Determine the next sector
		if seg.AdjacentSector != 0 {
			// A portal!
			adj = seg.AdjacentSegment.Sector
		} else if ri.CheckEntry {
			// A higher layer segment! If we hit it (entering), the next sector is THIS sector.
			adj = s
		} else {
			// We have a non-portal segment and we're not checking higher layer sector (Exit mode).
			// Check for lower layer overlap or grazing corner.

			// Nudge the intersection point slightly along the ray to see where we end up.
			// This handles grazing corners where we might clip a corner and exit instantly.
			nudgeX := (ri.Delta[0] / ri.Limit) * constants.IntersectEpsilon
			nudgeY := (ri.Delta[1] / ri.Limit) * constants.IntersectEpsilon

			testPoint := intersectionTest.To2D()
			testPoint[0] += nudgeX
			testPoint[1] += nudgeY

			// Check if we are overlapping with a lower layer sector (or any overlap).
			// OverlapAt(..., true) checks LowerLayers.
			adj = s.OverlapAt(testPoint, true)

			if adj == nil {
				// A solid wall! Next sector is nil.
			}
		}

		// Occlusion Checks (Floor/Ceiling)
		// Even if we hit a portal segment, the portal might be blocked by floor/ceiling differences.

		if intersectionTest[2] < s.Min[2] || intersectionTest[2] > s.Max[2] {
			// Outside sector bounds entirely? Occluded.
			// Treat as wall (hitSegment set, nextSector nil).
			adj = nil
		} else {
			i2d := intersectionTest.To2D()
			floorZ, ceilZ = s.ZAt(i2d)

			if intersectionTest[2] < floorZ-constants.IntersectEpsilon || intersectionTest[2] > ceilZ+constants.IntersectEpsilon {
				// Occluded by this sector's floor/ceiling
				adj = nil
			} else if !ri.CheckEntry && adj != nil {
				// If checking Exit, and we have a next sector, check its floor/ceiling too.
				// (If checkEntry is true, adj is s, so we already checked it).
				floorZ2, ceilZ2 = adj.ZAt(i2d)
				if intersectionTest[2] < floorZ2-constants.IntersectEpsilon || intersectionTest[2] > ceilZ2+constants.IntersectEpsilon {
					// Occluded by adjacent sector's floor/ceiling
					adj = nil
				}
			}
		}

		// Distance checks
		intersectionDistSq = intersectionTest.DistSq(&ri.Start)

		// 1. Is it better than current best?
		if intersectionDistSq >= ri.HitDistSq {
			continue
		}

		// 2. Is it ahead of previous hit?
		// Note: Using epsilon to be safe against floating point error.
		if intersectionDistSq <= ri.MinDistSq+constants.IntersectEpsilon {
			// LightSampler: if prevDistSq - intersectionDistSq > 0 { continue }
			// implies intersectionDistSq < prevDistSq -> continue.
			// We want intersectionDistSq > prevDistSq.
			continue
		}

		// We have a valid, better intersection.
		ri.HitDistSq = intersectionDistSq
		ri.HitPoint = intersectionTest
		ri.HitSegment = seg
		ri.NextSector = adj
	}
	return
}
