// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"strconv"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

func PickUpInventoryItem(ic *behaviors.InventoryCarrier, itemEntity ecs.Entity) {
	item := behaviors.GetInventoryItem(ic.ECS, itemEntity)
	if item == nil || !item.Active {
		return
	}
	player := behaviors.GetPlayer(ic.ECS, ic.Entity)

	for _, slot := range ic.Inventory {
		if !slot.ValidClasses.Contains(item.Class) {
			continue
		}
		if slot.Count.Now >= slot.Limit {
			if player != nil {
				player.Notices.Push("Can't pick up more " + item.Class)
			}
			return
		}
		toAdd := concepts.Min(item.Count.Now, slot.Limit-slot.Count.Now)
		slot.Count.Now += toAdd
		if player != nil {
			player.Notices.Push("Picked up " + strconv.Itoa(toAdd) + " " + item.Class)
		}
		//item.Count.Now -= toAdd
		// Disable all the entity components
		for _, c := range item.ECS.AllComponents(itemEntity) {
			if c != nil && !c.MultiAttachable() {
				c.Base().Active = false
			}
		}
	}
}

func LinkWeapons(ic *behaviors.InventoryCarrier, item *behaviors.InventoryItem) {
	col := ecs.ColumnFor[behaviors.WeaponClass](item.ECS, behaviors.WeaponClassCID)
	for i := range col.Cap() {
		wc := col.Value(i)
		if wc == nil || !wc.Active {
			continue
		}
		if ecs.GetNamed(wc.ECS, wc.Entity) {
		}

	}
}
