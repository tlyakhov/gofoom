// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"log"
	"math"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type WeaponInstantController struct {
	concepts.BaseController
	*behaviors.WeaponInstant
	Body *core.Body

	delta, isect concepts.Vector3
}

func init() {
	concepts.DbTypes().RegisterController(&WeaponInstantController{})
}

func (wc *WeaponInstantController) ComponentIndex() int {
	return behaviors.WeaponInstantComponentIndex
}

func (wc *WeaponInstantController) Methods() concepts.ControllerMethod {
	return concepts.ControllerAlways
}

func (wc *WeaponInstantController) Target(target concepts.Attachable) bool {
	wc.WeaponInstant = target.(*behaviors.WeaponInstant)
	wc.Body = core.BodyFromDb(wc.WeaponInstant.DB, wc.WeaponInstant.Entity)
	return wc.WeaponInstant.IsActive() && wc.Body.IsActive()
}

// This is similar to the code for lighting
func (wc *WeaponInstantController) Hit() *core.Selectable {
	// Since our sectors can be concave, we can't just go through the first portal we find,
	// we have to go through the NEAREST one. Use hitDist2 to keep track...
	hitDist2 := constants.MaxViewDistance * constants.MaxViewDistance
	dist2 := constants.MaxViewDistance * constants.MaxViewDistance
	idist2 := 0.0
	var s *core.Selectable
	p := &wc.Body.Pos.Now
	wc.delta[0] = math.Cos(wc.Body.Angle.Now * concepts.Deg2rad)
	wc.delta[1] = math.Sin(wc.Body.Angle.Now * concepts.Deg2rad)

	rayEnd := &concepts.Vector3{
		p[0] + wc.delta[0]*constants.MaxViewDistance,
		p[1] + wc.delta[1]*constants.MaxViewDistance,
		p[2]}

	sector := wc.Body.Sector()
	depth := 0 // We keep track of portaling depth to avoid infinite traversal in weird cases.
	for sector != nil {
		for _, b := range sector.Bodies {
			if !b.Active || b.Shadow == core.BodyShadowNone {
				continue
			}
			switch b.Shadow {
			case core.BodyShadowSphere:
				if concepts.IntersectLineSphere(p, rayEnd, b.Pos.Render, b.Size.Render[0]*0.5) {
					continue
				}
			case core.BodyShadowAABB:
				ext := &concepts.Vector3{b.Size.Render[0], b.Size.Render[0], b.Size.Render[1]}
				if !concepts.IntersectLineAABB(p, rayEnd, b.Pos.Render, ext) {
					continue
				}
			}
			idist2 = b.Pos.Render.Dist2(p)
			if idist2 < hitDist2 {
				s = core.SelectableFromBody(b)
				hitDist2 = idist2
			}
		}
		for _, seg := range sector.InternalSegments {
			// Find the intersection with this segment.
			if ok := seg.Intersect3D(p, rayEnd, &wc.isect); ok {
				idist2 := wc.isect.Dist2(p)
				if idist2 < hitDist2 {
					s = core.SelectableFromInternalSegment(seg)
					hitDist2 = idist2
				}
			}
		}

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
					s = core.SelectableFromWall(seg, core.SelectableMid)
					hitDist2 = idist2
				}
				continue
			}

			// Here, we know we have an intersected portal segment. It could still be occluding the light though, since the
			// bottom/top portions could be in the way.
			floorZ, ceilZ := sector.PointZ(concepts.DynamicRender, wc.isect.To2D())
			floorZ2, ceilZ2 := seg.AdjacentSegment.Sector.PointZ(concepts.DynamicRender, wc.isect.To2D())
			if wc.isect[2] < floorZ2 || wc.isect[2] < floorZ {
				idist2 = wc.isect.Dist2(p)
				if idist2 < hitDist2 {
					s = core.SelectableFromWall(seg, core.SelectableLow)
					hitDist2 = idist2
				}
				continue
			}
			if wc.isect[2] < floorZ2 || wc.isect[2] > ceilZ2 ||
				wc.isect[2] < floorZ || wc.isect[2] > ceilZ {
				idist2 = wc.isect.Dist2(p)
				if idist2 < hitDist2 {
					s = core.SelectableFromWall(seg, core.SelectableHi)
					hitDist2 = idist2
				}
				continue
			}

			// Get the square of the distance to the intersection (from the target point)
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

func (wc *WeaponInstantController) Always() {
	if !wc.FireNextFrame {
		return
	}
	wc.FireNextFrame = false
	s := wc.Hit()

	if s == nil {
		return
	}

	log.Printf("Weapon hit! %v, %v", s, wc.isect)
	// TODO: Cover all surface types
	// TODO: Keep track of a finite set of extra stages in a ring buffer.
	switch s.Type {
	case core.SelectableMid:
		u := wc.isect.To2D().Dist(&s.SectorSegment.P) / s.SectorSegment.Length
		v := 0.5
		es := &materials.ShaderStage{Texture: wc.MarkMaterial}
		es.SetDB(s.DB)
		es.Transform.SetIdentity()
		es.Transform.ScaleSelf(10)
		es.Transform[4] = -u * 10
		es.Transform[5] = -v * 10
		wc.ShaderStages = append(wc.ShaderStages, es)
		s.SectorSegment.Surface.ExtraStages = append(s.SectorSegment.Surface.ExtraStages, es)
	}
}
