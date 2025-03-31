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
	item := behaviors.GetInventoryItem(ic.Universe, itemEntity)
	if item == nil || !item.IsActive() {
		return
	}
	player := behaviors.GetPlayer(ic.Universe, ic.Entity)

	for _, e := range ic.Inventory {
		if e == 0 {
			continue
		}

		slot := behaviors.GetInventorySlot(ic.Universe, e)

		if slot == nil || slot.Class != item.Class {
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
		for _, c := range item.Universe.AllComponents(itemEntity) {
			if c != nil && !c.MultiAttachable() {
				c.Base().Flags &= ^ecs.ComponentActive
			}
		}
	}
}
