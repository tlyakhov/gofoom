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
	// Create a player if we don't have one
	if wc.DB.First(behaviors.PlayerComponentIndex) == nil {
		player := archetypes.CreatePlayerBody(wc.DB)
		playerBody := core.BodyFromDb(wc.DB, player)
		playerBody.Pos.Original = wc.Spawn.Spawn
		playerBody.Pos.Reset()
	}
}

func (wc *WorldController) proximity(sector *core.Sector, body *core.Body) {
	// Consider the case where the sector entity has a proximity
	// component that includes the body as a valid scripting source
	if p := behaviors.ProximityFromDb(wc.DB, sector.Entity); p != nil && p.IsActive() {
		if sector.Center.Dist2(&body.Pos.Now) < p.Range*p.Range {
			BodySectorScript(p.Scripts, body.Entity, sector.Entity)
		}
	}

	// Consider the case where the body entity has a proximity
	// component that includes the sector as a valid scripting source
	if p := behaviors.ProximityFromDb(wc.DB, body.Entity); p != nil && p.IsActive() {
		if sector.Center.Dist2(&body.Pos.Now) < p.Range*p.Range {
			BodySectorScript(p.Scripts, body.Entity, sector.Entity)
		}
	}
}

func (wc *WorldController) Always() {
	for _, c := range wc.DB.Components[core.BodyComponentIndex] {
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
