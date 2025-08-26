// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/ecs"
)

var flowerSlot *inventory.Slot
var gunSlot *inventory.Slot

func validateSpawn() {
	arena := ecs.ArenaByID(character.PlayerCID).(*ecs.Arena[character.Player, *character.Player])

	for i := range arena.Len() {
		player := arena.Value(i)
		if player == nil || !player.Spawn {
			continue
		}
		mobile := core.GetMobile(player.Entity)
		if mobile == nil {
			mobile = ecs.NewAttachedComponent(player.Entity, core.MobileCID).(*core.Mobile)
		}
		mobile.Elasticity = 0.1
		carrier := inventory.GetCarrier(player.Entity)
		if carrier == nil {
			carrier = ecs.NewAttachedComponent(player.Entity, inventory.CarrierCID).(*inventory.Carrier)
		}
		// TODO: What if the level creator has set up custom slots, or this is a
		// savegame?
		/*if ecs.CachedGeneratedComponent(u, &flowerSlot, "_PlayerInventoryFlower", inventory.SlotCID) {
			flowerSlot.Class = "Flower"
			flowerSlot.Limit = 5
			flowerSlot.Image = ecs.GetEntityByName("Pluk")
		}
		if ecs.CachedGeneratedComponent(u, &gunSlot, "_PlayerInventoryGun", inventory.SlotCID) {
			gunSlot.Class = "WeirdGun"
			gunSlot.Limit = 1
			gunSlot.Image = ecs.GetEntityByName("WeirdGun")
		}

		carrier.Inventory = nil
		carrier.Inventory.Set(flowerSlot.Entity)
		carrier.Inventory.Set(gunSlot.Entity)*/
	}
}
