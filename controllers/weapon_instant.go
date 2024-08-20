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
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/render/state"
)

type WeaponInstantController struct {
	ecs.BaseController
	*behaviors.WeaponInstant
	state.MaterialSampler
	Body *core.Body

	delta, isect, hit concepts.Vector3
	transform         concepts.Matrix2
}

func init() {
	ecs.Types().RegisterController(&WeaponInstantController{}, 100)
}

func (wc *WeaponInstantController) ComponentIndex() int {
	return behaviors.WeaponInstantComponentIndex
}

func (wc *WeaponInstantController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways
}

func (wc *WeaponInstantController) Target(target ecs.Attachable) bool {
	wc.WeaponInstant = target.(*behaviors.WeaponInstant)
	wc.Body = core.GetBody(wc.WeaponInstant.ECS, wc.WeaponInstant.Entity)
	return wc.WeaponInstant.IsActive() && wc.Body.IsActive()
}

// This is similar to the code for lighting
func (wc *WeaponInstantController) Cast() *core.Selectable {
	var s *core.Selectable
	wc.MaterialSampler.Config = &state.Config{ECS: wc.Body.ECS}
	wc.MaterialSampler.Ray = &state.Ray{Angle: wc.Body.Angle.Now}

	hitDist2 := constants.MaxViewDistance * constants.MaxViewDistance
	idist2 := 0.0
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
			if !b.Active || b.Entity == wc.Entity {
				continue
			}
			if ok := wc.InitializeRayBody(p, rayEnd, b); ok {
				wc.SampleMaterial(nil)
				if wc.MaterialSampler.Output[3] < 0.9 {
					continue
				}
				idist2 = b.Pos.Now.Dist2(p)
				if idist2 < hitDist2 {
					s = core.SelectableFromBody(b)
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
					s = core.SelectableFromInternalSegment(seg)
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
					s = core.SelectableFromWall(seg, core.SelectableMid)
					hitDist2 = idist2
					wc.hit = wc.isect
				}
				continue
			}

			// Here, we know we have an intersected portal segment. It could still be occluding the light though, since the
			// bottom/top portions could be in the way.
			floorZ, ceilZ := sector.ZAt(ecs.DynamicNow, wc.isect.To2D())
			floorZ2, ceilZ2 := seg.AdjacentSegment.Sector.ZAt(ecs.DynamicNow, wc.isect.To2D())
			if wc.isect[2] < floorZ2 || wc.isect[2] < floorZ {
				idist2 = wc.isect.Dist2(p)
				if idist2 < hitDist2 {
					s = core.SelectableFromWall(seg, core.SelectableLow)
					hitDist2 = idist2
					wc.hit = wc.isect
				}
				continue
			}
			if wc.isect[2] < floorZ2 || wc.isect[2] > ceilZ2 ||
				wc.isect[2] < floorZ || wc.isect[2] > ceilZ {
				idist2 = wc.isect.Dist2(p)
				if idist2 < hitDist2 {
					s = core.SelectableFromWall(seg, core.SelectableHi)
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

func (wc *WeaponInstantController) MarkSurface(s *core.Selectable, p *concepts.Vector2) (surf *materials.Surface, bottom, top float64) {
	switch s.Type {
	case core.SelectableHi:
		_, top = s.Sector.ZAt(ecs.DynamicNow, p)
		adj := s.SectorSegment.AdjacentSegment.Sector
		_, adjTop := adj.ZAt(ecs.DynamicNow, p)
		if adjTop <= top {
			bottom = adjTop
			surf = &s.SectorSegment.AdjacentSegment.HiSurface
		} else {
			bottom, top = top, adjTop
			surf = &s.SectorSegment.HiSurface
		}
	case core.SelectableLow:
		bottom, _ = s.Sector.ZAt(ecs.DynamicNow, p)
		adj := s.SectorSegment.AdjacentSegment.Sector
		adjBottom, _ := adj.ZAt(ecs.DynamicNow, p)
		if bottom <= adjBottom {
			top = adjBottom
			surf = &s.SectorSegment.AdjacentSegment.LoSurface
		} else {
			bottom, top = adjBottom, bottom
			surf = &s.SectorSegment.LoSurface
		}
	case core.SelectableMid:
		bottom, top = s.Sector.ZAt(ecs.DynamicNow, p)
		surf = &s.SectorSegment.Surface
	case core.SelectableInternalSegment:
		bottom, top = s.InternalSegment.Bottom, s.InternalSegment.Top
		surf = &s.InternalSegment.Surface
	}
	return
}

func (wc *WeaponInstantController) MarkSurfaceAndTransform(s *core.Selectable, transform *concepts.Matrix2) *materials.Surface {
	// Inverse of the size of bullet mark we want
	scale := 1.0 / wc.MarkSize
	// 3x2 transformation matrixes are composed of
	// the horizontal basis vector in slots [0] & [1], which we set to the
	// width of the segment, scaled
	// the vertical basis vector in slots [2] & [3], which we set to the
	// height of the segment
	// and finally the translation in slots [4] & [5], which we set to the
	// world position of the mark, relative to the segment
	switch s.Type {
	case core.SelectableHi:
		fallthrough
	case core.SelectableLow:
		fallthrough
	case core.SelectableMid:
		transform[concepts.MatBasis1X] = s.SectorSegment.Length
		transform[concepts.MatTransX] = -wc.hit.To2D().Dist(&s.SectorSegment.P)
	case core.SelectableInternalSegment:
		transform[concepts.MatBasis1X] = s.InternalSegment.Length
		transform[concepts.MatTransX] = -wc.hit.To2D().Dist(s.InternalSegment.A)
	}

	surf, bottom, top := wc.MarkSurface(s, wc.hit.To2D())
	// This is reversed because our UV coordinates go top->bottom
	transform[concepts.MatBasis2Y] = (bottom - top)
	transform[concepts.MatTransY] = -(wc.Body.Pos.Now[2] - top)

	transform[concepts.MatBasis1X] *= scale
	transform[concepts.MatBasis2Y] *= scale
	transform[concepts.MatTransX] *= scale
	transform[concepts.MatTransY] *= scale

	return surf
}

func (wc *WeaponInstantController) updateMarks(mark behaviors.WeaponMark) {
	wc.Marks.PushBack(mark)
	for wc.Marks.Len() > constants.MaxWeaponMarks {
		wm := wc.Marks.PopFront()
		for i, stage := range wm.Surface.ExtraStages {
			if stage != wm.ShaderStage {
				continue
			}
			wm.Surface.ExtraStages = append(wm.Surface.ExtraStages[:i], wm.Surface.ExtraStages[i+1:]...)
			break
		}
	}
}

func (wc *WeaponInstantController) Always() {
	if !wc.FireNextFrame {
		return
	}
	wc.FireNextFrame = false
	s := wc.Cast()

	if s == nil {
		return
	}

	// TODO: Account for bullet velocity travel time. Do this by calculating
	// time it would take to hit the thing and delaying the outcome? could be
	// buggy though if the object in question moves
	log.Printf("Weapon hit! %v[%v] at %v", s.Type, s.Entity, wc.hit.StringHuman())
	if s.Type == core.SelectableBody {
		// Push bodies away
		// TODO: Parameterize in WeaponInstant
		s.Body.Vel.Now.AddSelf(wc.delta.Mul(3))
		// Hurt anything alive
		if alive := behaviors.GetAlive(wc.Body.ECS, s.Body.Entity); alive != nil {
			// TODO: Parameterize in WeaponInstant
			alive.Hurt("Weapon "+s.Entity.String(), 5, 20)
		}
		// TODO: Death animations/entity deactivation
	} else if s.Type == core.SelectableSectorSegment ||
		s.Type == core.SelectableHi ||
		s.Type == core.SelectableLow ||
		s.Type == core.SelectableMid ||
		s.Type == core.SelectableInternalSegment {
		// Make a mark on walls

		// TODO: Include floors and ceilings
		es := &materials.ShaderStage{
			Texture:                wc.MarkMaterial,
			IgnoreSurfaceTransform: false,
			System:                 true}
		es.SetECS(s.ECS)
		surf := wc.MarkSurfaceAndTransform(s, &wc.transform)
		surf.ExtraStages = append(surf.ExtraStages, es)
		es.Transform.From(&surf.Transform.Now)
		es.Transform.AffineInverseSelf().MulSelf(&wc.transform)
		wc.updateMarks(behaviors.WeaponMark{
			ShaderStage: es,
			Surface:     surf,
		})
	}
}
