// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"strconv"
	"tlyakhov/gofoom/components/audio"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/ecs"
)

func PickUpInventoryItem(ic *inventory.Carrier, itemEntity ecs.Entity) {
	item := inventory.GetItem(itemEntity)
	if item == nil || !item.IsActive() {
		return
	}
	player := character.GetPlayer(ic.Entity)

	for _, e := range ic.Slots {
		if e == 0 {
			continue
		}

		slot := inventory.GetSlot(e)

		if slot == nil || slot.Class != item.Class {
			continue
		}
		if slot.Count.Now >= slot.Limit {
			if player != nil {
				player.Notices.Push("Can't pick up more " + item.Class)
			}
			return
		}
		toAdd := min(item.Count.Now, slot.Limit-slot.Count.Now)
		slot.Count.Now += toAdd
		if player != nil {
			player.Notices.Push("Picked up " + strconv.Itoa(toAdd) + " " + item.Class)
			audio.PlaySound(item.PickupSound, ic.Entity, "inventorypickup", audio.SoundPlayNormal)
		}
		//item.Count.Now -= toAdd
		// Disable all the entity components
		for _, c := range ecs.AllComponents(itemEntity) {
			if c != nil && !c.Shareable() {
				c.Base().Flags &= ^ecs.ComponentActive
			}
		}
	}
}
