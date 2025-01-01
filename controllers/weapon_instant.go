// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"log"
	"math"
	"math/rand/v2"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/render"
)

type WeaponInstantController struct {
	ecs.BaseController
	*behaviors.WeaponInstant

	Sampler render.MaterialSampler
	Slot    *behaviors.InventorySlot
	Carrier *behaviors.InventoryCarrier
	Class   *behaviors.WeaponClass
	Body    *core.Body

	delta, isect, hit concepts.Vector3
	transform         concepts.Matrix2
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &WeaponInstantController{} }, 100)
}

func (wc *WeaponInstantController) ComponentID() ecs.ComponentID {
	return behaviors.WeaponInstantCID
}

func (wc *WeaponInstantController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways
}

func (wc *WeaponInstantController) Target(target ecs.Attachable, e ecs.Entity) bool {
	wc.Entity = e
	wc.WeaponInstant = target.(*behaviors.WeaponInstant)
	wc.Class = behaviors.GetWeaponClass(wc.WeaponInstant.ECS, e)
	if wc.Class == nil || !wc.Class.IsActive() {
		return false
	}
	wc.Slot = behaviors.GetInventorySlot(wc.WeaponInstant.ECS, e)
	if wc.Slot == nil || !wc.Slot.IsActive() ||
		wc.Slot.Carrier == nil || !wc.Slot.Carrier.IsActive() {
		return false
	}
	// The source of our shot is the body attached to the inventory carrier
	wc.Body = core.GetBody(wc.WeaponInstant.ECS, wc.Slot.Carrier.Entity)
	return wc.WeaponInstant.IsActive() &&
		wc.Body != nil && wc.Body.IsActive() &&
		wc.Class != nil && wc.Class.IsActive()
}

// This is similar to the code for lighting
func (wc *WeaponInstantController) Cast() *selection.Selectable {
	var s *selection.Selectable

	angle := wc.Body.Angle.Now + (rand.Float64()-0.5)*wc.Class.Spread
	zSpread := (rand.Float64() - 0.5) * wc.Class.Spread

	wc.Sampler.Config = &render.Config{ECS: wc.Body.ECS}
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
			if !b.Active || b.Entity == wc.Body.Entity {
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

func (wc *WeaponInstantController) MarkSurface(s *selection.Selectable, p *concepts.Vector2) (surf *materials.Surface, bottom, top float64) {
	switch s.Type {
	case selection.SelectableHi:
		top = s.Sector.Top.ZAt(dynamic.DynamicNow, p)
		adj := s.SectorSegment.AdjacentSegment.Sector
		adjTop := adj.Top.ZAt(dynamic.DynamicNow, p)
		if adjTop <= top {
			bottom = adjTop
			surf = &s.SectorSegment.AdjacentSegment.HiSurface
		} else {
			bottom, top = top, adjTop
			surf = &s.SectorSegment.HiSurface
		}
	case selection.SelectableLow:
		bottom = s.Sector.Bottom.ZAt(dynamic.DynamicNow, p)
		adj := s.SectorSegment.AdjacentSegment.Sector
		adjBottom := adj.Bottom.ZAt(dynamic.DynamicNow, p)
		if bottom <= adjBottom {
			top = adjBottom
			surf = &s.SectorSegment.AdjacentSegment.LoSurface
		} else {
			bottom, top = adjBottom, bottom
			surf = &s.SectorSegment.LoSurface
		}
	case selection.SelectableMid:
		bottom, top = s.Sector.ZAt(dynamic.DynamicNow, p)
		surf = &s.SectorSegment.Surface
	case selection.SelectableInternalSegment:
		bottom, top = s.InternalSegment.Bottom, s.InternalSegment.Top
		surf = &s.InternalSegment.Surface
	}
	return
}

// TODO: This is more generally useful as a way to create transforms
// that map world-space onto texture space. This should be refactored to be part
// of anything with a surface
func (wc *WeaponInstantController) MarkSurfaceAndTransform(s *selection.Selectable, transform *concepts.Matrix2) *materials.Surface {
	// Inverse of the size of bullet mark we want
	scale := 1.0 / wc.Class.MarkSize
	// 3x2 transformation matrixes are composed of
	// the horizontal basis vector in slots [0] & [1], which we set to the
	// width of the segment, scaled
	// the vertical basis vector in slots [2] & [3], which we set to the
	// height of the segment
	// and finally the translation in slots [4] & [5], which we set to the
	// world position of the mark, relative to the segment
	switch s.Type {
	case selection.SelectableHi, selection.SelectableLow, selection.SelectableMid:
		transform[concepts.MatBasis1X] = s.SectorSegment.Length
		transform[concepts.MatTransX] = -wc.hit.To2D().Dist(&s.SectorSegment.P)
	case selection.SelectableInternalSegment:
		transform[concepts.MatBasis1X] = s.InternalSegment.Length
		transform[concepts.MatTransX] = -wc.hit.To2D().Dist(s.InternalSegment.A)
	}

	surf, bottom, top := wc.MarkSurface(s, wc.hit.To2D())
	// This is reversed because our UV coordinates go top->bottom
	transform[concepts.MatBasis2Y] = (bottom - top)
	transform[concepts.MatTransY] = -(wc.hit[2] - top)

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
	wc.FiredTimestamp = wc.ECS.Timestamp
	s := wc.Cast()

	if s == nil {
		return
	}

	// TODO: Account for bullet velocity travel time. Do this by calculating
	// time it would take to hit the thing and delaying the outcome? could be
	// buggy though if the object in question moves
	log.Printf("Weapon hit! %v[%v] at %v", s.Type, s.Entity, wc.hit.StringHuman(2))
	if s.Type == selection.SelectableBody {
		if mobile := core.GetMobile(s.Body.ECS, s.Body.Entity); mobile != nil {
			// Push bodies away
			// TODO: Parameterize in WeaponInstant
			mobile.Vel.Now.AddSelf(wc.delta.Mul(3))
		}
		// Hurt anything alive
		if alive := behaviors.GetAlive(wc.Body.ECS, s.Body.Entity); alive != nil {
			// TODO: Parameterize in WeaponInstant
			alive.Hurt("Weapon "+s.Entity.String(), wc.Class.Damage, 20)
		}
		// TODO: Death animations/entity deactivation
	} else if s.Type == selection.SelectableSectorSegment ||
		s.Type == selection.SelectableHi ||
		s.Type == selection.SelectableLow ||
		s.Type == selection.SelectableMid ||
		s.Type == selection.SelectableInternalSegment {
		// Make a mark on walls

		// TODO: Include floors and ceilings
		es := &materials.ShaderStage{
			Material:               wc.Class.MarkMaterial,
			IgnoreSurfaceTransform: false,
			System:                 true}
		es.OnAttach(s.ECS)
		es.Construct(nil)
		es.Flags = 0
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
