// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
)

type WorldController struct {
	concepts.BaseController
	*core.Spawn
}

func init() {
	concepts.DbTypes().RegisterController(&WorldController{})
}

func (wc *WorldController) ComponentIndex() int {
	return core.SpawnComponentIndex
}

func (wc *WorldController) Priority() int {
	return 70
}

func (wc *WorldController) Methods() concepts.ControllerMethod {
	return concepts.ControllerLoaded | concepts.ControllerAlways
}

func (wc *WorldController) Target(target concepts.Attachable) bool {
	wc.Spawn = target.(*core.Spawn)
	return wc.IsActive()
}

func (wc *WorldController) Loaded() {
	var player *behaviors.Player
	// Create a player if we don't have one
	if player = wc.DB.First(behaviors.PlayerComponentIndex).(*behaviors.Player); player != nil {
		playerBody := core.BodyFromDb(wc.DB, player.Entity)
		playerBody.Elasticity = 0.1
	} else {
		entity := archetypes.CreatePlayerBody(wc.DB)
		playerBody := core.BodyFromDb(wc.DB, entity)
		playerBody.Pos.Original = wc.Spawn.Spawn
		playerBody.Pos.ResetToOriginal()
		player = wc.DB.First(behaviors.PlayerComponentIndex).(*behaviors.Player)
	}
	player.Inventory = make([]*behaviors.InventorySlot, 2)
	player.Inventory[0] = &behaviors.InventorySlot{
		Image:        player.DB.GetEntityByName("Pluk"),
		Limit:        5,
		ValidClasses: make(concepts.Set[string]),
	}
	player.Inventory[0].ValidClasses.Add("Flower")
	player.Inventory[1] = &behaviors.InventorySlot{
		Image:        133,
		Limit:        1,
		ValidClasses: make(concepts.Set[string]),
	}
	player.Inventory[1].ValidClasses.Add("WeirdGun")
}

func (wc *WorldController) proximity(sector *core.Sector, body *core.Body) {
	// Consider the case where the sector entity has a proximity
	// component that includes the body as a valid scripting source
	if p := behaviors.ProximityFromDb(wc.DB, sector.Entity); p != nil && p.IsActive() {
		if sector.Center.Dist2(&body.Pos.Now) < p.Range*p.Range {
			BodySectorScript(p.Scripts, body, sector)
		}
	}

	// Consider the case where the body entity has a proximity
	// component that includes the sector as a valid scripting source
	if p := behaviors.ProximityFromDb(wc.DB, body.Entity); p != nil && p.IsActive() {
		if sector.Center.Dist2(&body.Pos.Now) < p.Range*p.Range {
			BodySectorScript(p.Scripts, body, sector)
		}
	}
}

func (wc *WorldController) Always() {
	for _, c := range wc.DB.AllOfType(core.BodyComponentIndex) {
		body := c.(*core.Body)
		if !body.Active || body.Sector() == nil {
			continue
		}
		for _, pvs := range body.Sector().PVS {
			wc.proximity(pvs, body)
		}
	}
}

func DefaultMaterial(db *concepts.EntityComponentDB) concepts.Entity {
	entity := db.GetEntityByName("Default Material")
	if entity != 0 {
		return entity
	}

	// Otherwise try a random one?
	return db.First(materials.LitComponentIndex).GetEntity()
}
