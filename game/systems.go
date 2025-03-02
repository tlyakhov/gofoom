// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
)

var flowerSlot *behaviors.InventorySlot
var gunSlot *behaviors.InventorySlot

func validateSpawn(db *ecs.ECS) {
	col := db.Column(behaviors.PlayerCID).(*ecs.Column[behaviors.Player, *behaviors.Player])

	for i := range col.Len() {
		player := col.Value(i)
		if player == nil || !player.Spawn {
			continue
		}
		mobile := core.GetMobile(db, player.Entity)
		if mobile == nil {
			mobile = db.NewAttachedComponent(player.Entity, core.MobileCID).(*core.Mobile)
		}
		mobile.Elasticity = 0.1
		carrier := behaviors.GetInventoryCarrier(db, player.Entity)
		if carrier == nil {
			carrier = db.NewAttachedComponent(player.Entity, behaviors.InventoryCarrierCID).(*behaviors.InventoryCarrier)
		}
		// TODO: What if the level creator has set up custom slots, or this is a
		// savegame?
		/*if ecs.CachedGeneratedComponent(db, &flowerSlot, "_PlayerInventoryFlower", behaviors.InventorySlotCID) {
			flowerSlot.Class = "Flower"
			flowerSlot.Limit = 5
			flowerSlot.Image = db.GetEntityByName("Pluk")
		}
		if ecs.CachedGeneratedComponent(db, &gunSlot, "_PlayerInventoryGun", behaviors.InventorySlotCID) {
			gunSlot.Class = "WeirdGun"
			gunSlot.Limit = 1
			gunSlot.Image = db.GetEntityByName("WeirdGun")
		}

		carrier.Inventory = nil
		carrier.Inventory.Set(flowerSlot.Entity)
		carrier.Inventory.Set(gunSlot.Entity)*/
	}
}
