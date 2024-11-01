// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

//go:generate go run github.com/dmarkham/enumer -type=InventoryIndex -json
type InventoryIndex int

const (
	InventoryWeirdGun InventoryIndex = iota
	InventoryFlower
	InventoryCount
)

//entity := archetypes.CreatePlayerBody(db)
//playerBody := core.GetBody(db, entity)
/*if spawn != nil {
	playerBody.Pos.Spawn = spawn.Spawn
	playerBody.Pos.ResetToSpawn()
}*/
//player = behaviors.GetPlayer(db, entity)

/*playerBody := core.GetBody(db, player.Entity)
playerBody.Elasticity = 0.1

player.Inventory = make([]*behaviors.InventorySlot, InventoryCount)
slot := &behaviors.InventorySlot{
	Image:        player.ECS.GetEntityByName("Pluk"),
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
player.Inventory[InventoryWeirdGun] = slot*/
