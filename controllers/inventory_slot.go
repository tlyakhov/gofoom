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
	WeaponClass   *behaviors.WeaponClass
	WeaponInstant *behaviors.WeaponInstant
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
	isc.WeaponInstant = behaviors.GetWeaponInstant(isc.ECS, e)
	return isc.InventorySlot != nil && isc.InventorySlot.IsActive()
}

func (isc *InventorySlotController) Recalculate() {
	if isc.WeaponClass != nil && isc.WeaponInstant == nil {
		isc.WeaponInstant = isc.ECS.NewAttachedComponent(isc.Entity, behaviors.WeaponInstantCID).(*behaviors.WeaponInstant)
		isc.WeaponInstant.System = true
	}
}
