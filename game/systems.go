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

func validateSpawn(u *ecs.Universe) {
	col := u.Column(character.PlayerCID).(*ecs.Column[character.Player, *character.Player])

	for i := range col.Len() {
		player := col.Value(i)
		if player == nil || !player.Spawn {
			continue
		}
		mobile := core.GetMobile(u, player.Entity)
		if mobile == nil {
			mobile = u.NewAttachedComponent(player.Entity, core.MobileCID).(*core.Mobile)
		}
		mobile.Elasticity = 0.1
		carrier := inventory.GetCarrier(u, player.Entity)
		if carrier == nil {
			carrier = u.NewAttachedComponent(player.Entity, inventory.CarrierCID).(*inventory.Carrier)
		}
		// TODO: What if the level creator has set up custom slots, or this is a
		// savegame?
		/*if ecs.CachedGeneratedComponent(u, &flowerSlot, "_PlayerInventoryFlower", inventory.SlotCID) {
			flowerSlot.Class = "Flower"
			flowerSlot.Limit = 5
			flowerSlot.Image = u.GetEntityByName("Pluk")
		}
		if ecs.CachedGeneratedComponent(u, &gunSlot, "_PlayerInventoryGun", inventory.SlotCID) {
			gunSlot.Class = "WeirdGun"
			gunSlot.Limit = 1
			gunSlot.Image = u.GetEntityByName("WeirdGun")
		}

		carrier.Inventory = nil
		carrier.Inventory.Set(flowerSlot.Entity)
		carrier.Inventory.Set(gunSlot.Entity)*/
	}
}
