// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"math/rand/v2"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/render"
)

// This is similar to the code for lighting
func (wc *WeaponController) Cast() *selection.Selectable {
	var s *selection.Selectable

	angle := wc.Body.Angle.Now + (rand.Float64()-0.5)*wc.Class.Spread
	zSpread := (rand.Float64() - 0.5) * wc.Class.Spread

	wc.Sampler.Config = &render.Config{}
	wc.Sampler.Ray = &render.Ray{Angle: angle}

	hitDist2 := constants.MaxViewDistance * constants.MaxViewDistance
	idist2 := 0.0
	p := &wc.Body.Pos.Now
	wc.delta[0] = math.Cos(angle * concepts.Deg2rad)
	wc.delta[1] = math.Sin(angle * concepts.Deg2rad)
	wc.delta[2] = math.Sin(zSpread * concepts.Deg2rad)

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
				idist2 = b.Pos.Now.Dist2(p)
				if idist2 < hitDist2 {
					s = selection.SelectableFromBody(b)
					hitDist2 = idist2
					wc.hit = b.Pos.Now
				}
			}
		}
		for _, seg := range sector.InternalSegments {
			// Find the intersection with this segment.
			if ok := seg.Intersect3D(p, rayEnd, &wc.isect); ok {
				idist2 = wc.isect.Dist2(p)
				if idist2 < hitDist2 {
					s = selection.SelectableFromInternalSegment(seg)
					hitDist2 = idist2
					wc.hit = wc.isect
				}
			}
		}

		// Since our sectors can be concave, we can't just go through the first portal we find,
		// we have to go through the NEAREST one. Use dist2 to keep track...
		dist2 := constants.MaxViewDistance * constants.MaxViewDistance
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
				idist2 = wc.isect.Dist2(p)
				if idist2 < hitDist2 {
					s = selection.SelectableFromWall(seg, selection.SelectableMid)
					hitDist2 = idist2
					wc.hit = wc.isect
				}
				continue
			}

			// Here, we know we have an intersected portal segment. It could still be occluding the light though, since the
			// bottom/top portions could be in the way.
			floorZ, ceilZ := sector.ZAt(dynamic.DynamicNow, wc.isect.To2D())
			floorZ2, ceilZ2 := seg.AdjacentSegment.Sector.ZAt(dynamic.DynamicNow, wc.isect.To2D())
			if wc.isect[2] < floorZ2 || wc.isect[2] < floorZ {
				idist2 = wc.isect.Dist2(p)
				if idist2 < hitDist2 {
					s = selection.SelectableFromWall(seg, selection.SelectableLow)
					hitDist2 = idist2
					wc.hit = wc.isect
				}
				continue
			}
			if wc.isect[2] < floorZ2 || wc.isect[2] > ceilZ2 ||
				wc.isect[2] < floorZ || wc.isect[2] > ceilZ {
				idist2 = wc.isect.Dist2(p)
				if idist2 < hitDist2 {
					s = selection.SelectableFromWall(seg, selection.SelectableHi)
					hitDist2 = idist2
					wc.hit = wc.isect
				}
				continue
			}

			// Get the square of the distance to the intersection
			idist2 := wc.isect.Dist2(p)
			if idist2-dist2 > constants.IntersectEpsilon {
				// If the current intersection point is farther than one we
				// already have for this sector, we have a concavity, body, or
				// internal segment. Keep looking.
				continue
			}

			// We're in the clear! Move to the next adjacent sector.
			dist2 = idist2
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
