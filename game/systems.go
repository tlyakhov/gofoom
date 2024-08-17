// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

//go:generate go run github.com/dmarkham/enumer -type=InventoryIndex -json
type InventoryIndex int

const (
	InventoryWeirdGun InventoryIndex = iota
	InventoryFlower
	InventoryCount
)

func setupPlayer() {
	var player *behaviors.Player
	// Create a player if we don't have one
	if player = db.First(behaviors.PlayerComponentIndex).(*behaviors.Player); player != nil {
		playerBody := core.BodyFromDb(db, player.Entity)
		playerBody.Elasticity = 0.1
	} else {
		spawn := db.First(core.SpawnComponentIndex).(*core.Spawn)
		entity := archetypes.CreatePlayerBody(db)
		playerBody := core.BodyFromDb(db, entity)
		if spawn != nil {
			playerBody.Pos.Original = spawn.Spawn
			playerBody.Pos.ResetToOriginal()
		}
		player = db.First(behaviors.PlayerComponentIndex).(*behaviors.Player)
	}

	player.Inventory = make([]*behaviors.InventorySlot, InventoryCount)
	slot := &behaviors.InventorySlot{
		Image:        player.DB.GetEntityByName("Pluk"),
		Limit:        5,
		ValidClasses: make(concepts.Set[string]),
	}
	slot.ValidClasses.Add("Flower")
	player.Inventory[InventoryFlower] = slot
	slot = &behaviors.InventorySlot{
		Image:        133,
		Limit:        1,
		ValidClasses: make(concepts.Set[string]),
	}
	slot.ValidClasses.Add("WeirdGun")
	player.Inventory[InventoryWeirdGun] = slot
}
