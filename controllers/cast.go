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
	var req core.CastRequest

	sampler.Config = &render.Config{}
	sampler.Ray = ray

	limitSq := ray.Limit * ray.Limit
	lastBoundaryDistSq := -1.0

	// Initialize Ray State
	req.Ray = ray
	req.IgnoreSegment = nil

	depth := 0 // We keep track of portaling depth to avoid infinite traversal in weird cases.
	for sector != nil {
		req.HitDistSq = limitSq
		req.HitSegment = nil
		req.NextSector = nil
		s = nil // Reset selection

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
					if idistSq < req.HitDistSq && idistSq > lastBoundaryDistSq {
						s = selection.SelectableFromBody(b)
						req.HitDistSq = idistSq
						hit = b.Pos.Now
					}
				}
			}
		}
		for _, seg := range sector.InternalSegments {
			// Find the intersection with this segment.
			if ok := seg.Intersect3D(&ray.Start, &ray.End, &isect); ok {
				idistSq := isect.DistSq(&ray.Start)
				if idistSq < req.HitDistSq && idistSq > lastBoundaryDistSq {
					s = selection.SelectableFromInternalSegment(seg)
					req.HitDistSq = idistSq
					hit = isect
				}
			}
		}

		// Check sector boundaries (Exit)
		req.MinDistSq = lastBoundaryDistSq
		req.CheckEntry = false
		sector.IntersectRay(&req)

		// Check higher layer sectors (Entry)
		bestEntry := false
		bestHit := req.CastResponse

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
			req.CheckEntry = true
			if overlap.IntersectRay(&req); req.HitSegment != nil {
				bestHit = req.CastResponse
				bestEntry = true
			}
		}

		if bestHit.HitSegment == nil {
			// Nothing hit in this sector, and no boundary found
			return
		}

		// We hit a boundary closer than any object
		s = nil
		hit = bestHit.HitPoint

		if bestHit.NextSector == nil {
			// Solid Wall Hit (or Occluded Portal)
			if bestEntry {
				// Hitting the outside of a higher layer sector
				s = selection.SelectableFromWall(bestHit.HitSegment, selection.SelectableMid)
			} else if bestHit.HitSegment.AdjacentSector != 0 {
				// Portal Occlusion
				floorZ, _ := sector.ZAt(bestHit.HitPoint.To2D())
				floorZ2, _ := bestHit.HitSegment.AdjacentSegment.Sector.ZAt(bestHit.HitPoint.To2D())
				if bestHit.HitPoint[2] < floorZ2 || bestHit.HitPoint[2] < floorZ {
					s = selection.SelectableFromWall(bestHit.HitSegment, selection.SelectableLow)
				} else {
					s = selection.SelectableFromWall(bestHit.HitSegment, selection.SelectableHi)
				}
			} else {
				// Solid Wall
				s = selection.SelectableFromWall(bestHit.HitSegment, selection.SelectableMid)
			}
			return // Hit wall, return
		}
		// Traverse
		lastBoundaryDistSq = bestHit.HitDistSq
		sector = bestHit.NextSector
		depth++
		if depth > constants.MaxPortals {
			return
		}
	}

	return
}
