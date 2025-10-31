// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"math/rand/v2"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/render"
)

// This is similar to the code for lighting
// TODO: Implement inner sector casting
func (wc *WeaponController) Cast() *selection.Selectable {
	var s *selection.Selectable

	angle := wc.Body.Angle.Now + (rand.Float64()-0.5)*wc.Class.Spread
	pitchSpread := (rand.Float64() - 0.5) * wc.Class.Spread

	wc.Sampler.Config = &render.Config{}
	wc.Sampler.Ray = &render.Ray{Angle: angle}

	hitDistSq := constants.MaxViewDistance * constants.MaxViewDistance
	idistSq := 0.0
	p := &wc.Body.Pos.Now
	wc.delta[0] = math.Cos(angle * concepts.Deg2rad)
	wc.delta[1] = math.Sin(angle * concepts.Deg2rad)

	// TODO: All bodies should probably be able to pitch
	if p := character.GetPlayer(wc.Body.Entity); p != nil {
		wc.delta[2] = math.Sin((p.Pitch + pitchSpread) * concepts.Deg2rad)
	} else {
		wc.delta[2] = math.Sin(pitchSpread * concepts.Deg2rad)
	}

	rayEnd := &concepts.Vector3{
		p[0] + wc.delta[0]*constants.MaxViewDistance,
		p[1] + wc.delta[1]*constants.MaxViewDistance,
		p[2] + wc.delta[2]*constants.MaxViewDistance}

	sector := wc.Body.Sector()
	depth := 0 // We keep track of portaling depth to avoid infinite traversal in weird cases.
	for sector != nil {
		for _, b := range sector.Bodies {
			if !b.IsActive() || b.Entity == wc.Body.Entity {
				continue
			}
			if ok := wc.Sampler.InitializeRayBody(p, rayEnd, b); ok {
				wc.Sampler.SampleMaterial(nil)
				if wc.Sampler.Output[3] < 0.9 {
					continue
				}
				idistSq = b.Pos.Now.DistSq(p)
				if idistSq < hitDistSq {
					s = selection.SelectableFromBody(b)
					hitDistSq = idistSq
					wc.hit = b.Pos.Now
				}
			}
		}
		for _, seg := range sector.InternalSegments {
			// Find the intersection with this segment.
			if ok := seg.Intersect3D(p, rayEnd, &wc.isect); ok {
				idistSq = wc.isect.DistSq(p)
				if idistSq < hitDistSq {
					s = selection.SelectableFromInternalSegment(seg)
					hitDistSq = idistSq
					wc.hit = wc.isect
				}
			}
		}

		// Since our sectors can be concave, we can't just go through the first portal we find,
		// we have to go through the NEAREST one. Use distSq to keep track...
		distSq := constants.MaxViewDistance * constants.MaxViewDistance
		var next *core.Sector
		for _, seg := range sector.Segments {
			// Segment facing backwards from our ray? skip it.
			if wc.delta[0]*seg.Normal[0]+wc.delta[1]*seg.Normal[1] > 0 {
				continue
			}

			// Find the intersection with this segment.
			//log.Printf("Testing %v<->%v, sector %v, seg %v:", p.StringHuman(), rayEnd.StringHuman(), sector.String(), seg.P.StringHuman())
			if ok := seg.Intersect3D(p, rayEnd, &wc.isect); !ok {
				continue // No intersection, skip it!
			}

			if seg.AdjacentSector == 0 {
				idistSq = wc.isect.DistSq(p)
				if idistSq < hitDistSq {
					s = selection.SelectableFromWall(seg, selection.SelectableMid)
					hitDistSq = idistSq
					wc.hit = wc.isect
				}
				continue
			}

			// Here, we know we have an intersected portal segment. It could still be occluding the light though, since the
			// bottom/top portions could be in the way.
			floorZ, ceilZ := sector.ZAt(wc.isect.To2D())
			floorZ2, ceilZ2 := seg.AdjacentSegment.Sector.ZAt(wc.isect.To2D())
			if wc.isect[2] < floorZ2 || wc.isect[2] < floorZ {
				idistSq = wc.isect.DistSq(p)
				if idistSq < hitDistSq {
					s = selection.SelectableFromWall(seg, selection.SelectableLow)
					hitDistSq = idistSq
					wc.hit = wc.isect
				}
				continue
			}
			if wc.isect[2] < floorZ2 || wc.isect[2] > ceilZ2 ||
				wc.isect[2] < floorZ || wc.isect[2] > ceilZ {
				idistSq = wc.isect.DistSq(p)
				if idistSq < hitDistSq {
					s = selection.SelectableFromWall(seg, selection.SelectableHi)
					hitDistSq = idistSq
					wc.hit = wc.isect
				}
				continue
			}

			// Get the square of the distance to the intersection
			idistSq := wc.isect.DistSq(p)
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
			return s
		}
		sector = next
	}

	return s
}
