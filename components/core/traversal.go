// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package core

import (
	"log"
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

const LogDebug = false

type CastResponse struct {
	HitSegment *SectorSegment
	HitPoint   concepts.Vector3
	HitDistSq  float64
	NextSector *Sector
}

type CastRequest struct {
	// Inputs
	*concepts.Ray

	IgnoreSegment *Segment
	MinDistSq     float64
	CheckEntry    bool
	Debug         bool

	// Input/Output
	CastResponse
}

// IntersectRay finds the nearest intersection of a ray with the sector's segments.
// It handles portal transitions, higher layer checks, and grazing ray edge cases.
//
// Arguments:
//   - ri: A RayIntersection struct containing ray definition, traversal state, and output fields.
//     The function will update ri.HitSegment, ri.HitPoint, ri.HitDistSq, and ri.NextSector.
func (s *Sector) IntersectRay(req *CastRequest) {
	debug := LogDebug && req.Debug

	// Pre-declare variables for loop
	var intersectionTest concepts.Vector3
	var adj *Sector
	var floorZ, ceilZ, floorZ2, ceilZ2, intersectionDistSq float64

	for _, seg := range s.Segments {
		if debug {
			log.Printf("    Checking segment [%v]-[%v]\n", seg.P.Render.StringHuman(), seg.Next.P.Render.StringHuman())
		}

		// Don't intersect with the segment we started on
		if &seg.Segment == req.IgnoreSegment {
			continue
		}

		// When checking higher layers (Entry check), we ignore portals in that layer.
		// We only care about entering the sector via a solid boundary (conceptually).
		if req.CheckEntry && seg.AdjacentSector != 0 {
			if debug {
				log.Printf("    Ignoring portal to %v in higher layer sector.", seg.AdjacentSector)
			}
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
		dot := req.Delta[0]*seg.Normal[0] + req.Delta[1]*seg.Normal[1]
		normalFacing := dot > 0

		if req.CheckEntry != normalFacing {
			if debug {
				log.Printf("    Ignoring segment [or behind]. higher layer? = %v, normal? = %v. Dot: %v", req.CheckEntry, normalFacing, dot)
			}
			continue
		}

		// Find the intersection with this segment.
		if !seg.Intersect3D(&req.Start, &req.End, &intersectionTest) {
			if debug {
				log.Printf("    No intersection\n")
			}
			continue
		}

		if debug {
			log.Printf("    Intersection = [%v]\n", intersectionTest.StringHuman(2))
		}

		// Determine the next sector
		if seg.AdjacentSector != 0 {
			// A portal!
			adj = seg.AdjacentSegment.Sector
			if debug {
				log.Printf("    Portal to %v", adj.Entity)
			}
		} else if req.CheckEntry {
			// A higher layer segment! If we hit it (entering), the next sector is THIS sector.
			adj = s
			if debug {
				log.Printf("    Higher layer to %v", adj.Entity)
			}
		} else {
			// We have a non-portal segment and we're not checking higher layer sector (Exit mode).
			// Check for lower layer overlap or grazing corner.

			// Nudge the intersection point slightly along the ray to see where we end up.
			// This handles grazing corners where we might clip a corner and exit instantly.
			nudgeX := req.Delta[0] * constants.IntersectEpsilon
			nudgeY := req.Delta[1] * constants.IntersectEpsilon

			testPoint := intersectionTest.To2D()
			testPoint[0] += nudgeX
			testPoint[1] += nudgeY

			// Check if we are overlapping with a lower layer sector (or any overlap).
			// OverlapAt(..., true) checks LowerLayers.
			adj = s.OverlapAt(testPoint, true)

			if adj == nil {
				if debug {
					log.Printf("    Occluded behind wall seg %v|%v\n", seg.P.Render.StringHuman(), seg.Next.P.Render.StringHuman())
				}
				// A solid wall! Next sector is nil.
			} else {
				if debug {
					log.Printf("    Out to %v", adj.Entity)
				}
			}
		}

		// Occlusion Checks (Floor/Ceiling)
		// Even if we hit a portal segment, the portal might be blocked by floor/ceiling differences.

		/*if intersectionTest[2] < s.Min[2] || intersectionTest[2] > s.Max[2] {
			if debug {
				log.Printf("    Occluded by sector min/max %v - %v\n", seg.P.Render.StringHuman(), seg.Next.P.Render.StringHuman())
			}
			// Outside sector bounds entirely? Occluded.
			// Treat as wall (hitSegment set, nextSector nil).
			adj = nil
		} else {*/
		i2d := intersectionTest.To2D()
		floorZ, ceilZ = s.ZAt(i2d)

		if intersectionTest[2] < floorZ-constants.IntersectEpsilon || intersectionTest[2] > ceilZ+constants.IntersectEpsilon {
			if debug {
				log.Printf("    Occluded by floor/ceiling gap: %v - %v\n", seg.P.Render.StringHuman(), seg.Next.P.Render.StringHuman())
			}
			// Occluded by this sector's floor/ceiling
			adj = nil
		} else if !req.CheckEntry && adj != nil {
			// If checking Exit, and we have a next sector, check its floor/ceiling too.
			// (If checkEntry is true, adj is s, so we already checked it).
			floorZ2, ceilZ2 = adj.ZAt(i2d)
			if intersectionTest[2] < floorZ2-constants.IntersectEpsilon || intersectionTest[2] > ceilZ2+constants.IntersectEpsilon {
				if debug {
					log.Printf("    Occluded by floor/ceiling gap: %v - %v\n", seg.P.Render.StringHuman(), seg.Next.P.Render.StringHuman())
				}
				// Occluded by adjacent sector's floor/ceiling
				adj = nil
			}
		}
		//}

		// Distance checks
		intersectionDistSq = intersectionTest.DistSq(&req.Start)

		// 1. Is it better than current best?
		if intersectionDistSq >= req.HitDistSq {
			continue
		}

		// 2. Is it ahead of previous hit?
		// Note: Using epsilon to be safe against floating point error.
		// We allow hits at roughly the same distance to handle corner cases where we exit one sector
		// and immediately exit the next (shared vertex).
		if intersectionDistSq < req.MinDistSq-constants.IntersectEpsilon {
			if debug {
				log.Printf("    Found intersection point before the previous sector: %v < %v\n", math.Sqrt(intersectionDistSq), math.Sqrt(req.MinDistSq))
			}
			continue
		}

		if debug {
			log.Printf("    Found a portal to %v without impediment at dist %v.\n", adj, math.Sqrt(intersectionDistSq))
		}

		// We have a valid, better intersection.
		req.HitDistSq = intersectionDistSq
		req.HitPoint = intersectionTest
		req.HitSegment = seg
		req.NextSector = adj
	}
}
