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
	sampler.Config = &render.Config{}
	sampler.Ray = ray

	limitSq := ray.Limit * ray.Limit
	lastBoundaryDistSq := -1.0

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
		bSeg, bPoint, bDist, bNext := sector.IntersectRay(&ray.Start, &ray.End, &ray.Delta, ray.Limit, nil, lastBoundaryDistSq, hitDistSq, false)

		// Check higher layer sectors (Entry)
		bestEntry := false
		for _, e := range sector.HigherLayers {
			if e == 0 {
				continue
			}
			overlap := core.GetSector(e)
			if overlap == nil {
				continue
			}
			// Use the best boundary distance found so far (bDist or hitDistSq)
			currentBest := hitDistSq
			if bSeg != nil {
				currentBest = bDist
			}

			hSeg, hPoint, hDist, hNext := overlap.IntersectRay(&ray.Start, &ray.End, &ray.Delta, ray.Limit, nil, lastBoundaryDistSq, currentBest, true)
			if hSeg != nil {
				bSeg = hSeg
				bPoint = hPoint
				bDist = hDist
				bNext = hNext
				bestEntry = true
			}
		}

		if bSeg != nil {
			// We hit a boundary closer than any object
			s = nil
			hit = bPoint

			if bNext == nil {
				// Solid Wall Hit (or Occluded Portal)
				if bestEntry {
					// Hitting the outside of a higher layer sector
					s = selection.SelectableFromWall(bSeg, selection.SelectableMid)
				} else if bSeg.AdjacentSector != 0 {
					// Portal Occlusion
					floorZ, _ := sector.ZAt(bPoint.To2D())
					floorZ2, _ := bSeg.AdjacentSegment.Sector.ZAt(bPoint.To2D())
					if bPoint[2] < floorZ2 || bPoint[2] < floorZ {
						s = selection.SelectableFromWall(bSeg, selection.SelectableLow)
					} else {
						s = selection.SelectableFromWall(bSeg, selection.SelectableHi)
					}
				} else {
					// Solid Wall
					s = selection.SelectableFromWall(bSeg, selection.SelectableMid)
				}
				return // Hit wall, return
			} else {
				// Traverse
				lastBoundaryDistSq = bDist
				sector = bNext
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
