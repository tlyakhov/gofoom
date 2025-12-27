// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/ecs"
)

type InventorySlotController struct {
	ecs.BaseController
	*inventory.Slot
	WeaponClass *inventory.WeaponClass
	Weapon      *inventory.Weapon
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &InventorySlotController{} }, 75)
}

func (isc *InventorySlotController) ComponentID() ecs.ComponentID {
	return inventory.SlotCID
}

func (isc *InventorySlotController) Methods() ecs.ControllerMethod {
	return ecs.ControllerPrecompute
}

func (isc *InventorySlotController) EditorPausedMethods() ecs.ControllerMethod {
	return ecs.ControllerPrecompute
}

func (isc *InventorySlotController) Target(target ecs.Component, e ecs.Entity) bool {
	isc.Entity = e
	isc.Slot = target.(*inventory.Slot)
	isc.WeaponClass = inventory.GetWeaponClass(e)
	isc.Weapon = inventory.GetWeapon(e)
	return isc.Slot != nil && isc.Slot.IsActive()
}

func (isc *InventorySlotController) Precompute() {
	if isc.WeaponClass != nil && isc.Weapon == nil {
		isc.Weapon = ecs.NewAttachedComponent(isc.Entity, inventory.WeaponCID).(*inventory.Weapon)
		isc.Weapon.Flags |= ecs.ComponentInternal
	}
}
