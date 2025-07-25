// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math/rand/v2"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type InventoryItemController struct {
	ecs.BaseController
	*inventory.Item
	body *core.Body

	autoProximity        *behaviors.Proximity
	autoPlayerTargetable *behaviors.PlayerTargetable
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &InventoryItemController{} }, 75)
}

func (iic *InventoryItemController) ComponentID() ecs.ComponentID {
	return inventory.ItemCID
}

func (iic *InventoryItemController) Methods() ecs.ControllerMethod {
	return ecs.ControllerRecalculate
}

func (iic *InventoryItemController) EditorPausedMethods() ecs.ControllerMethod {
	return ecs.ControllerRecalculate
}

func (iic *InventoryItemController) Target(target ecs.Attachable, e ecs.Entity) bool {
	iic.Entity = e
	iic.Item = target.(*inventory.Item)
	if iic.Item == nil || !iic.Item.IsActive() {
		return false
	}
	iic.body = core.GetBody(iic.Entity)
	return true
}

func (iic *InventoryItemController) cacheAutoProximity() {
	if ecs.CachedGeneratedComponent(&iic.autoProximity, "_InventoryItemAutoProximity", behaviors.ProximityCID) {
		iic.autoProximity.Hysteresis = 0
		iic.autoProximity.InRange.Code = `
				if p := character.GetPlayer(body.Entity); p != nil {
					p.HoveringTargets.Add(onEntity)
				}`

		ecs.ActAllControllersOneEntity(iic.autoProximity.Entity, ecs.ControllerRecalculate)
	}
}

func (iic *InventoryItemController) cacheAutoTargetable() {
	if ecs.CachedGeneratedComponent(&iic.autoPlayerTargetable, "_InventoryItemAutoTargetable", behaviors.PlayerTargetableCID) {
		iic.autoPlayerTargetable.Frob.Code = `
			if carrier == nil || body == nil { return }
			controllers.PickUpInventoryItem(carrier, body.Entity)`
		iic.autoPlayerTargetable.Message = `Pick up {{with ecs_Named .TargetableEntity}}{{.Name}}{{else}}item{{end}}`
		ecs.ActAllControllersOneEntity(iic.autoPlayerTargetable.Entity, ecs.ControllerRecalculate)
	}
}

func (iic *InventoryItemController) Recalculate() {
	if iic.body != nil && (iic.Flags&inventory.ItemBounce != 0) {
		a := iic.body.Pos.NewAnimation()
		a.TweeningFunc = dynamic.EaseInOut2
		a.End[2] = 5
		a.Duration = 1000
		// Make inventory items bounce differently for variety
		a.Percent = rand.Float64()
	}

	if iic.Flags&inventory.ItemAutoProximity != 0 {
		iic.cacheAutoProximity()
		p := behaviors.GetProximity(iic.Entity)
		if p != nil && p != iic.autoProximity {
			ecs.DetachComponent(behaviors.ProximityCID, iic.Entity)
			p = nil
		}
		if p == nil {
			var a ecs.Attachable = iic.autoProximity
			ecs.Attach(behaviors.ProximityCID, iic.Entity, &a)
		}
	}
	if iic.Flags&inventory.ItemAutoPlayerTargetable != 0 {
		iic.cacheAutoTargetable()
		pt := behaviors.GetPlayerTargetable(iic.Entity)
		if pt != nil && pt != iic.autoPlayerTargetable {
			ecs.DetachComponent(behaviors.PlayerTargetableCID, iic.Entity)
			pt = nil
		}
		if pt == nil {
			var a ecs.Attachable = iic.autoPlayerTargetable
			ecs.Attach(behaviors.PlayerTargetableCID, iic.Entity, &a)
		}
	}
}
