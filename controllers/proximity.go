// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

type ProximityController struct {
	ecs.BaseController
	*behaviors.Proximity
}

func init() {
	ecs.Types().RegisterController(&ProximityController{}, 100)
}

func (pc *ProximityController) ComponentID() ecs.ComponentID {
	return behaviors.ProximityCID
}

func (pc *ProximityController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways
}

func (pc *ProximityController) Target(target ecs.Attachable) bool {
	pc.Proximity = target.(*behaviors.Proximity)
	return pc.IsActive()
}

func (pc *ProximityController) checkPlayer(entity ecs.Entity) bool {
	if !pc.RequiresPlayerAction {
		return true
	}
	if player := behaviors.GetPlayer(pc.ECS, entity); player != nil && player.ActionPressed {
		return true
	}
	return false
}

func (pc *ProximityController) fire(body *core.Body, sector *core.Sector) {
	pc.Firing = true
	if pc.LastFired+int64(pc.Hysteresis) > pc.ECS.Timestamp {
		return
	}
	pc.LastFired = pc.ECS.Timestamp
	for _, script := range pc.Scripts {
		script.Vars["proximityEntity"] = pc.Entity
		script.Vars["body"] = body
		script.Vars["sector"] = sector
		script.Act()
	}
}

func (pc *ProximityController) sectorBodies(proximityEntity ecs.Entity, sector *core.Sector, pos *concepts.Vector3) {
	for _, body := range sector.Bodies {
		if !pc.checkPlayer(body.Entity) {
			continue
		}
		if pos.Dist2(&body.Pos.Now) < pc.Range*pc.Range {
			if ps := core.GetSector(pc.ECS, proximityEntity); ps != nil {
				pc.fire(body, ps)
			} else {
				pc.fire(body, sector)
			}
		}
	}
}
func (pc *ProximityController) proximity(proximityEntity ecs.Entity) {
	if sector := core.GetSector(pc.ECS, proximityEntity); sector != nil {
		for _, pvs := range sector.PVS {
			pc.sectorBodies(proximityEntity, pvs, &sector.Center)
		}
		return
	}
	if body := core.GetBody(pc.ECS, proximityEntity); body != nil && body.SectorEntity != 0 {
		container := body.Sector()
		for _, sector := range container.PVS {
			if pc.ActsOnSectors && sector.Center.Dist2(&body.Pos.Now) < pc.Range*pc.Range {
				pc.fire(body, sector)
			}
			pc.sectorBodies(proximityEntity, sector, &body.Pos.Now)
		}
	}
}

func (pc *ProximityController) Always() {
	pc.Firing = false

	// Is the target itself a body or sector?
	pc.proximity(pc.Entity)

	for entity := range pc.Entities {
		pc.proximity(entity)
	}
}
