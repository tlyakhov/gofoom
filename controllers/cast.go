// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"log"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/render"
)

const LogDebug = false

// TODO: Add ability to filter what we select
func Cast(ray *concepts.Ray, sector *core.Sector, source ecs.Entity, ignoreBodies bool) (s *selection.Selectable, hit concepts.Vector3) {
	var sampler render.MaterialSampler
	var isect concepts.Vector3
	var req core.CastRequest

	// Initialize state
	sampler.Config = &render.Config{}
	sampler.Ray = ray
	if LogDebug {
		req.Debug = character.GetPlayer(source) != nil
		if req.Debug {
			log.Printf("Cast START: %v, starting sector %v", source, sector.Entity)
		}
	}
	req.Ray = ray
	req.IgnoreSegment = nil
	req.MinDistSq = -1.0

	s = nil // Reset selection
	limitSq := ray.Limit * ray.Limit
	depth := 0 // We keep track of portaling depth to avoid infinite traversal in weird cases.
	for sector != nil {
		req.HitDistSq = limitSq
		req.HitSegment = nil
		req.NextSector = nil

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
					if idistSq < req.HitDistSq && idistSq > req.MinDistSq {
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
				if idistSq < req.HitDistSq && idistSq > req.MinDistSq {
					s = selection.SelectableFromInternalSegment(seg)
					req.HitDistSq = idistSq
					hit = isect
				}
			}
		}

		// Check sector boundaries (Exit)
		req.CheckEntry = false
		sector.IntersectRay(&req)

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
			overlap.IntersectRay(&req)
		}
		bestHit := req.CastResponse

		if bestHit.HitSegment == nil {
			// Nothing hit in this sector, and no boundary found
			if LogDebug && req.Debug {
				log.Printf("Nothing hit in this sector, and no boundary found")
			}
			return
		}

		// We hit a boundary closer than any object
		s = nil
		hit = bestHit.HitPoint

		if bestHit.NextSector == nil {
			// Hit solid wall or occluded portal
			switch {
			case bestHit.HitPortal < 0:
				s = selection.SelectableFromWall(bestHit.HitSegment, selection.SelectableLow)
			case bestHit.HitPortal > 0:
				s = selection.SelectableFromWall(bestHit.HitSegment, selection.SelectableHi)
			default:
				s = selection.SelectableFromWall(bestHit.HitSegment, selection.SelectableMid)
			}
			return // Stop
		}
		// Traverse
		req.MinDistSq = bestHit.HitDistSq
		sector = bestHit.NextSector
		depth++
		if depth > constants.MaxPortals {
			return
		}
	}

	return
}
