// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/audio"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/concepts"
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

	delta, hit     concepts.Vector3
	transform      concepts.Matrix2
	bodyController BodyController
	markController MarkMakerController
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

func (w *WeaponController) newState(s inventory.WeaponState) {
	//log.Printf("Weapon %v changing from state %v->%v after %vms", w.Entity, w.State, s, (ecs.Simulation.SimTimestamp-w.LastStateTimestamp)/1_000_000)
	w.State = s
	w.LastStateTimestamp = ecs.Simulation.SimTimestamp
	p := w.Class.Params[w.State]
	if p.Sound != 0 {
		audio.PlaySound(p.Sound, w.Body.Entity, "weapon "+s.String(), audio.SoundPlayNormal)
	}
}

func weaponIdle(wc *WeaponController) {
	if wc.Intent == inventory.WeaponFire {
		wc.newState(inventory.WeaponFiring)
		return
	}
}

func weaponUnholstering(wc *WeaponController) {
	if wc.stateCompleted() {
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
		wc.fireWeaponProjectile(projectile)
	}
}

func weaponCooling(wc *WeaponController) {
	if wc.stateCompleted() {
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

func (wc *WeaponController) stateCompleted() bool {
	return ecs.Simulation.SimTimestamp-wc.LastStateTimestamp >= concepts.MillisToNanos(wc.Class.Params[wc.State].Time)
}
