// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/render"
)

// This is similar to the code for lighting
// TODO: Implement inner sector casting
// TODO: Add ability to filter what we select
func Cast(ray *concepts.Ray, sector *core.Sector, source ecs.Entity, ignoreBodies bool) (s *selection.Selectable, hit concepts.Vector3) {
	var sampler render.MaterialSampler
	var isect concepts.Vector3
	var ri core.RayIntersection

	sampler.Config = &render.Config{}
	sampler.Ray = ray

	limitSq := ray.Limit * ray.Limit
	lastBoundaryDistSq := -1.0

	// Initialize Ray State
	ri.Ray = ray
	ri.IgnoreSegment = nil

	depth := 0 // We keep track of portaling depth to avoid infinite traversal in weird cases.
	for sector != nil {
		hitDistSq := limitSq // Best hit in this sector so far
		s = nil              // Reset selection

		if !ignoreBodies {
			for _, b := range sector.Bodies {
				if !b.IsActive() || b.Entity == source {
					continue
				}
				if ok := sampler.InitializeRayBody(&ray.Start, &ray.End, b); ok {
					sampler.SampleMaterial(nil)
					if sampler.Output[3] < 0.9 {
						continue
					}
					idistSq := b.Pos.Now.DistSq(&ray.Start)
					if idistSq < hitDistSq && idistSq > lastBoundaryDistSq {
						s = selection.SelectableFromBody(b)
						hitDistSq = idistSq
						hit = b.Pos.Now
					}
				}
			}
		}
		for _, seg := range sector.InternalSegments {
			// Find the intersection with this segment.
			if ok := seg.Intersect3D(&ray.Start, &ray.End, &isect); ok {
				idistSq := isect.DistSq(&ray.Start)
				if idistSq < hitDistSq && idistSq > lastBoundaryDistSq {
					s = selection.SelectableFromInternalSegment(seg)
					hitDistSq = idistSq
					hit = isect
				}
			}
		}

		// Check sector boundaries (Exit)
		// We use hitDistSq as maxDistSq to ensure we only find boundaries closer than any object hit
		ri.MinDistSq = lastBoundaryDistSq
		ri.MaxDistSq = hitDistSq
		ri.CheckEntry = false
		sector.IntersectRay(&ri)

		// Check higher layer sectors (Entry)
		bestEntry := false
		// We need to store the "best result" from the normal exit check, because we're about to reuse 'ri' for entry checks.
		// However, we can just let 'ri' accumulate the best result.
		// If we run IntersectRay on an overlap, and it finds a closer hit, it updates ri.
		// If it doesn't, ri remains unchanged? No, IntersectRay resets fields.
		// So we must manually manage the "best found" candidate.

		// Let's store the best candidate in 'bestRI' (a copy of the result).
		// Since we can't easily deep copy without allocation, and we want to avoid allocations...
		// We will stick to the manual variables for the BEST candidate found so far.
		// But wait, the user asked "Why are bSeg, bPoint, bDist, and bNext necessary? Can't they be reused from RayIntersection?"
		// The answer is: IntersectRay resets them. But we can change that behavior or work around it.
		//
		// Alternative: RayIntersection holds the "current best".
		// Modify IntersectRay to NOT reset if it doesn't find a better one?
		// "ri.HitDistSq = ri.MaxDistSq" is set at the start.
		//
		// Use a second RayIntersection struct for the probe?
		// No, let's just use temp variables for the *current* probe, and keep the best in `ri`.
		// But `IntersectRay` takes `ri` as input/output.
		//
		// Let's use `ri` as the "best result holder".
		// When we check overlaps, we need to pass a struct that has the inputs (Ray, etc) but outputs to... where?
		// If we reuse `ri`, and `IntersectRay` fails to find a hit, it clears `ri.HitSegment`.
		// This clobbers our previous best hit.
		//
		// So we really do need `bSeg` etc. OR a second struct.
		// Given the user constraint "Can't they be reused", maybe they imply we should just use `ri` and manage it carefully.
		//
		// If we save the "best" hit in `ri`, and we want to check another sector...
		// We can't pass `ri` to `overlap.IntersectRay` without risking clobbering.
		// UNLESS we change `IntersectRay` to only update if it finds something BETTER?
		// `IntersectRay` currently does: `ri.HitSegment = nil`.
		//
		// Let's stick to using `bSeg` etc variables because they are effectively registers.
		// BUT the user asked to remove them.
		// "Why are bSeg, bPoint, bDist, and bNext necessary? Can't they be reused from RayIntersection?"
		//
		// Let's try to reuse `ri` by saving/restoring or using a temporary struct? No that's worse.
		//
		// Actually, `ri` is small. We can just copy it.
		// bestRI := ri
		// ... check overlaps with `testRI` ...
		// if testRI.HitSegment != nil && testRI.HitDistSq < bestRI.HitDistSq { bestRI = testRI }
		//
		// But `ri` has a pointer to Ray. Copying the struct is cheap (shallow copy).
		//
		// Let's implement that pattern.

		bestRI := ri
		// Force the initial 'best' to have HitSegment = nil if the first check failed
		// (IntersectRay already does this, so bestRI holds the result of the first check)

		// Check higher layer sectors (Entry)
		for _, e := range sector.HigherLayers {
			if e == 0 {
				continue
			}
			overlap := core.GetSector(e)
			if overlap == nil {
				continue
			}

			// We want to check this overlap.
			// Input: MaxDistSq should be the current best distance found.
			// If bestRI found something, use bestRI.HitDistSq. Else use limit.
			currentBestDist := hitDistSq
			if bestRI.HitSegment != nil {
				currentBestDist = bestRI.HitDistSq
			}

			// Setup a test RI
			testRI := ri // Copy inputs (Ray pointer, ignore segment, etc)
			testRI.MaxDistSq = currentBestDist
			testRI.CheckEntry = true

			overlap.IntersectRay(&testRI)

			if testRI.HitSegment != nil {
				bestRI = testRI
				bestEntry = true
			}
		}

		if bestRI.HitSegment != nil {
			// We hit a boundary closer than any object
			s = nil
			hit = bestRI.HitPoint

			if bestRI.NextSector == nil {
				// Solid Wall Hit (or Occluded Portal)
				if bestEntry {
					// Hitting the outside of a higher layer sector
					s = selection.SelectableFromWall(bestRI.HitSegment, selection.SelectableMid)
				} else if bestRI.HitSegment.AdjacentSector != 0 {
					// Portal Occlusion
					floorZ, _ := sector.ZAt(bestRI.HitPoint.To2D())
					floorZ2, _ := bestRI.HitSegment.AdjacentSegment.Sector.ZAt(bestRI.HitPoint.To2D())
					if bestRI.HitPoint[2] < floorZ2 || bestRI.HitPoint[2] < floorZ {
						s = selection.SelectableFromWall(bestRI.HitSegment, selection.SelectableLow)
					} else {
						s = selection.SelectableFromWall(bestRI.HitSegment, selection.SelectableHi)
					}
				} else {
					// Solid Wall
					s = selection.SelectableFromWall(bestRI.HitSegment, selection.SelectableMid)
				}
				return // Hit wall, return
			} else {
				// Traverse
				lastBoundaryDistSq = bestRI.HitDistSq
				sector = bestRI.NextSector
				depth++
				if depth > constants.MaxPortals {
					return
				}
				continue
			}
		}

		// If we hit an object and no boundary was closer
		if s != nil {
			return
		}

		// Nothing hit in this sector, and no boundary found
		return
	}

	return
}
