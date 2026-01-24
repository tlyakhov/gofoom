// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"log"
	"tlyakhov/gofoom/components/audio"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/render"
)

type WeaponController struct {
	ecs.BaseController
	*inventory.Weapon

	Sampler render.MaterialSampler
	Slot    *inventory.Slot
	Class   *inventory.WeaponClass
	Body    *core.Body

	delta, hit concepts.Vector3
	transform  concepts.Matrix2
}

var weaponFuncs = [inventory.WeaponStateCount]func(*WeaponController){}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &WeaponController{} }, 100)
	weaponFuncs[inventory.WeaponIdle] = weaponIdle
	weaponFuncs[inventory.WeaponUnholstering] = weaponUnholstering
	weaponFuncs[inventory.WeaponFiring] = weaponFiring
	weaponFuncs[inventory.WeaponCooling] = weaponCooling
	weaponFuncs[inventory.WeaponReloading] = weaponReloading
	weaponFuncs[inventory.WeaponHolstering] = weaponHolstering
}

func (wc *WeaponController) ComponentID() ecs.ComponentID {
	return inventory.WeaponCID
}

func (wc *WeaponController) Methods() ecs.ControllerMethod {
	return ecs.ControllerFrame
}

func (wc *WeaponController) Target(target ecs.Component, e ecs.Entity) bool {
	wc.Entity = e
	wc.Weapon = target.(*inventory.Weapon)
	wc.Class = inventory.GetWeaponClass(e)
	if wc.Class == nil || !wc.Class.IsActive() {
		return false
	}
	wc.Slot = inventory.GetSlot(e)
	if wc.Slot == nil || !wc.Slot.IsActive() ||
		wc.Slot.Carrier == nil || !wc.Slot.Carrier.IsActive() {
		return false
	}
	// The source of our shot is the body attached to the inventory carrier
	wc.Body = core.GetBody(wc.Slot.Carrier.Entity)
	return wc.Weapon.IsActive() &&
		wc.Body != nil && wc.Body.IsActive() &&
		wc.Class != nil && wc.Class.IsActive()
}

func (wc *WeaponController) MarkSurface(s *selection.Selectable, p *concepts.Vector2) (surf *materials.Surface, bottom, top float64) {
	switch s.Type {
	case selection.SelectableHi:
		top = s.Sector.Top.ZAt(p)
		adj := s.SectorSegment.AdjacentSegment.Sector
		adjTop := adj.Top.ZAt(p)
		if adjTop <= top {
			bottom = adjTop
			surf = &s.SectorSegment.AdjacentSegment.HiSurface
		} else {
			bottom, top = top, adjTop
			surf = &s.SectorSegment.HiSurface
		}
	case selection.SelectableLow:
		bottom = s.Sector.Bottom.ZAt(p)
		adj := s.SectorSegment.AdjacentSegment.Sector
		adjBottom := adj.Bottom.ZAt(p)
		if bottom <= adjBottom {
			top = adjBottom
			surf = &s.SectorSegment.AdjacentSegment.LoSurface
		} else {
			bottom, top = adjBottom, bottom
			surf = &s.SectorSegment.LoSurface
		}
	case selection.SelectableMid:
		bottom, top = s.Sector.ZAt(p)
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
func (wc *WeaponController) MarkSurfaceAndTransform(s *selection.Selectable, transform *concepts.Matrix2) *materials.Surface {
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
		transform[concepts.MatTransX] = -wc.hit.To2D().Dist(&s.SectorSegment.P.Render)
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

func (wc *WeaponController) updateMarks(mark inventory.WeaponMark) {
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

func (w *WeaponController) newState(s inventory.WeaponState) {
	log.Printf("Weapon %v changing from state %v->%v after %vms", w.Entity, w.State, s, (ecs.Simulation.SimTimestamp-w.LastStateTimestamp)/1_000_000)
	w.State = s
	w.LastStateTimestamp = ecs.Simulation.SimTimestamp
	p := w.Class.Params[w.State]
	if p.Sound != 0 {
		audio.PlaySound(p.Sound, w.Body.Entity, "weapon "+s.String(), false)
	}
}

func weaponIdle(wc *WeaponController) {
	if wc.Intent == inventory.WeaponFire {
		wc.newState(inventory.WeaponFiring)
		return
	}
}

func weaponUnholstering(wc *WeaponController) {
	if wc.StateCompleted() {
		wc.newState(inventory.WeaponIdle)
	}
	if wc.Intent == inventory.WeaponHolstered {
		wc.newState(inventory.WeaponHolstering)
	}
}

func weaponFiring(wc *WeaponController) {
	if wc.Intent != inventory.WeaponFire {
		return
	}
	// TODO: This should be a parameter somewhere, whether to reset the intent
	// or not.
	wc.Intent = inventory.WeaponHeld
	wc.newState(inventory.WeaponCooling)
	if instant := inventory.GetWeaponClassInstant(wc.Entity); instant != nil {
		wc.fireWeaponInstant(instant)
	}
	if projectile := inventory.GetWeaponClassProjectile(wc.Entity); projectile != nil {

	}
}

func weaponCooling(wc *WeaponController) {
	if wc.StateCompleted() {
		if wc.Intent == inventory.WeaponFire {
			wc.newState(inventory.WeaponFiring)
		} else {
			wc.newState(inventory.WeaponIdle)
		}
		return
	}
	if wc.Intent == inventory.WeaponHolstered {
		wc.newState(inventory.WeaponHolstering)
	}
}
func weaponReloading(wc *WeaponController) {
}
func weaponHolstering(wc *WeaponController) {
}

func (wc *WeaponController) Frame() {
	// Run our gun state machine
	weaponFuncs[wc.State](wc)
}
