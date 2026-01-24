// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/ecs"
)

type InventoryCarrierController struct {
	ecs.BaseController
	*inventory.Carrier
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &InventoryCarrierController{} }, 75)
}

func (icc *InventoryCarrierController) ComponentID() ecs.ComponentID {
	return inventory.CarrierCID
}

func (icc *InventoryCarrierController) Methods() ecs.ControllerMethod {
	return ecs.ControllerFrame
}

func (icc *InventoryCarrierController) EditorPausedMethods() ecs.ControllerMethod {
	return ecs.ControllerFrame
}

func (icc *InventoryCarrierController) Target(target ecs.Component, e ecs.Entity) bool {
	icc.Entity = e
	icc.Carrier = target.(*inventory.Carrier)
	return icc.Carrier != nil && icc.Carrier.IsActive()
}

func (icc *InventoryCarrierController) Frame() {
	for _, e := range icc.Carrier.Slots {
		if e == 0 {
			continue
		}
		slot := inventory.GetSlot(e)
		if slot == nil {
			continue
		}
		// TODO: Put this into an Carrier controller
		if slot.Carrier != icc.Carrier {
			slot.Carrier = icc.Carrier
		}

		//log.Printf("Slot %v, count spawn: %v, count now: %v", e, slot.Count.Spawn, slot.Count.Now)
		if slot.Count.Now <= 0 {
			continue
		}
		if w := inventory.GetWeapon(e); w != nil {
			icc.Carrier.SelectedWeapon = e
		}
	}
}
