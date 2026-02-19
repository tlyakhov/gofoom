// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

// IntersectRay finds the nearest intersection of a ray with the sector's segments.
// It handles portal transitions, higher layer checks, and grazing ray edge cases.
//
// Arguments:
//   - rayStart, rayEnd: The start and end points of the ray segment to test against.
//   - rayDelta: The direction vector of the ray (End - Start).
//   - rayLength: The length of the ray (used for normalization in epsilon nudging).
//   - ignoreSegment: A segment to ignore (usually the one the ray started from) to prevent self-intersection.
//   - minDistSq: The squared distance from the ray start that must be exceeded (exclusive) for a hit to be valid.
//     Used to prevent re-detecting intersections at the start point or "behind" the ray.
//   - maxDistSq: The maximum squared distance (exclusive) for a hit to be considered.
//     Serves as the initial "best found distance".
//   - checkEntry: If true, checks for ray *entering* the sector (normal facing opposite to ray, dot > 0).
//     If false, checks for ray *exiting* the sector (normal facing same as ray, dot <= 0).
//
// Returns:
//   - hitSegment: The segment that was hit, or nil if no valid intersection found.
//   - hitPoint: The point of intersection.
//   - hitDistSq: The squared distance to the intersection point.
//   - nextSector: The sector to transition to (portal adjacent, higher layer, or overlap), or nil if it's a solid wall.
func (s *Sector) IntersectRay(rayStart, rayEnd, rayDelta *concepts.Vector3, rayLength float64, ignoreSegment *Segment, minDistSq, maxDistSq float64, checkEntry bool) (hitSegment *SectorSegment, hitPoint concepts.Vector3, hitDistSq float64, nextSector *Sector) {
	hitDistSq = maxDistSq
	var intersectionTest concepts.Vector3

	// Pre-declare variables for loop
	var adj *Sector
	var floorZ, ceilZ, floorZ2, ceilZ2, intersectionDistSq float64

	for _, seg := range s.Segments {
		// Don't intersect with the segment we started on
		if &seg.Segment == ignoreSegment {
			continue
		}

		// When checking higher layers (Entry check), we ignore portals in that layer.
		// We only care about entering the sector via a solid boundary (conceptually).
		if checkEntry && seg.AdjacentSector != 0 {
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
		dot := rayDelta[0]*seg.Normal[0] + rayDelta[1]*seg.Normal[1]
		normalFacing := dot > 0

		if checkEntry != normalFacing {
			continue
		}

		// Find the intersection with this segment.
		if !seg.Intersect3D(rayStart, rayEnd, &intersectionTest) {
			continue
		}

		// Determine the next sector
		if seg.AdjacentSector != 0 {
			// A portal!
			adj = seg.AdjacentSegment.Sector
		} else if checkEntry {
			// A higher layer segment! If we hit it (entering), the next sector is THIS sector.
			adj = s
		} else {
			// We have a non-portal segment and we're not checking higher layer sector (Exit mode).
			// Check for lower layer overlap or grazing corner.

			// Nudge the intersection point slightly along the ray to see where we end up.
			// This handles grazing corners where we might clip a corner and exit instantly.
			nudgeX := (rayDelta[0] / rayLength) * constants.IntersectEpsilon
			nudgeY := (rayDelta[1] / rayLength) * constants.IntersectEpsilon

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
			} else if !checkEntry && adj != nil {
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
		intersectionDistSq = intersectionTest.DistSq(rayStart)

		// 1. Is it better than current best?
		if intersectionDistSq >= hitDistSq {
			continue
		}

		// 2. Is it ahead of previous hit?
		// Note: Using epsilon to be safe against floating point error.
		if intersectionDistSq <= minDistSq + constants.IntersectEpsilon {
			// LightSampler: if prevDistSq - intersectionDistSq > 0 { continue }
			// implies intersectionDistSq < prevDistSq -> continue.
			// We want intersectionDistSq > prevDistSq.
			continue
		}

		// We have a valid, better intersection.
		hitDistSq = intersectionDistSq
		hitPoint = intersectionTest
		hitSegment = seg
		nextSector = adj
	}
	return
}
