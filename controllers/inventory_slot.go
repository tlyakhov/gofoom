// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/ecs"
)

type InventorySlotController struct {
	ecs.BaseController
	*behaviors.InventorySlot
	WeaponClass *behaviors.WeaponClass
	Weapon      *behaviors.Weapon
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &InventorySlotController{} }, 75)
}

func (isc *InventorySlotController) ComponentID() ecs.ComponentID {
	return behaviors.InventorySlotCID
}

func (isc *InventorySlotController) Methods() ecs.ControllerMethod {
	return ecs.ControllerRecalculate
}

func (isc *InventorySlotController) EditorPausedMethods() ecs.ControllerMethod {
	return ecs.ControllerRecalculate
}

func (isc *InventorySlotController) Target(target ecs.Attachable, e ecs.Entity) bool {
	isc.Entity = e
	isc.InventorySlot = target.(*behaviors.InventorySlot)
	isc.WeaponClass = behaviors.GetWeaponClass(isc.ECS, e)
	isc.Weapon = behaviors.GetWeapon(isc.ECS, e)
	return isc.InventorySlot != nil && isc.InventorySlot.IsActive()
}

func (isc *InventorySlotController) Recalculate() {
	if isc.WeaponClass != nil && isc.Weapon == nil {
		isc.Weapon = isc.ECS.NewAttachedComponent(isc.Entity, behaviors.WeaponCID).(*behaviors.Weapon)
		isc.Weapon.System = true
	}
}
