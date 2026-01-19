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

	hitDistSq := ray.Limit * ray.Limit
	idistSq := 0.0

	depth := 0 // We keep track of portaling depth to avoid infinite traversal in weird cases.
	for sector != nil {
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
					idistSq = b.Pos.Now.DistSq(&ray.Start)
					if idistSq < hitDistSq {
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
				idistSq = isect.DistSq(&ray.Start)
				if idistSq < hitDistSq {
					s = selection.SelectableFromInternalSegment(seg)
					hitDistSq = idistSq
					hit = isect
				}
			}
		}

		// Since our sectors can be concave, we can't just go through the first portal we find,
		// we have to go through the NEAREST one. Use distSq to keep track...
		distSq := ray.Limit * ray.Limit
		var next *core.Sector
		for _, seg := range sector.Segments {
			// Segment facing backwards from our ray? skip it.
			if ray.Delta[0]*seg.Normal[0]+ray.Delta[1]*seg.Normal[1] > 0 {
				continue
			}

			// Find the intersection with this segment.
			//log.Printf("Testing %v<->%v, sector %v, seg %v:", ray.Start.StringHuman(), rayEnd.StringHuman(), sector.String(), seg.ray.Start.StringHuman())
			if ok := seg.Intersect3D(&ray.Start, &ray.End, &isect); !ok {
				continue // No intersection, skip it!
			}

			if seg.AdjacentSector == 0 {
				idistSq = isect.DistSq(&ray.Start)
				if idistSq < hitDistSq {
					s = selection.SelectableFromWall(seg, selection.SelectableMid)
					hitDistSq = idistSq
					hit = isect
				}
				continue
			}

			// Here, we know we have an intersected portal segment. It could still be occluding the light though, since the
			// bottom/top portions could be in the way.
			floorZ, ceilZ := sector.ZAt(isect.To2D())
			floorZ2, ceilZ2 := seg.AdjacentSegment.Sector.ZAt(isect.To2D())
			if isect[2] < floorZ2 || isect[2] < floorZ {
				idistSq = isect.DistSq(&ray.Start)
				if idistSq < hitDistSq {
					s = selection.SelectableFromWall(seg, selection.SelectableLow)
					hitDistSq = idistSq
					hit = isect
				}
				continue
			}
			if isect[2] < floorZ2 || isect[2] > ceilZ2 ||
				isect[2] < floorZ || isect[2] > ceilZ {
				idistSq = isect.DistSq(&ray.Start)
				if idistSq < hitDistSq {
					s = selection.SelectableFromWall(seg, selection.SelectableHi)
					hitDistSq = idistSq
					hit = isect
				}
				continue
			}

			// Get the square of the distance to the intersection
			idistSq := isect.DistSq(&ray.Start)
			if idistSq-distSq > constants.IntersectEpsilon {
				// If the current intersection point is farther than one we
				// already have for this sector, we have a concavity, body, or
				// internal segment. Keep looking.
				continue
			}

			// We're in the clear! Move to the next adjacent sector.
			distSq = idistSq
			next = seg.AdjacentSegment.Sector
		}
		depth++
		if s != nil || next == nil || depth > constants.MaxPortals { // Avoid infinite looping.
			return
		}
		sector = next
	}

	return
}
