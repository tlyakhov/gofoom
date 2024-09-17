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

	flags    behaviors.ProximityFlags
	onEntity ecs.Entity
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

func (pc *ProximityController) isEntityPlayerAndActing(entity ecs.Entity) bool {
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
		script.Vars["proximity"] = pc.Proximity
		script.Vars["onEntity"] = pc.onEntity
		script.Vars["body"] = body
		script.Vars["sector"] = sector
		script.Vars["flags"] = pc.flags
		script.Act()
	}
}

func (pc *ProximityController) sectorBodies(sector *core.Sector, pos *concepts.Vector3) {
	for _, body := range sector.Bodies {
		if (pc.flags&behaviors.ProximityTargetsBody) != 0 &&
			!pc.isEntityPlayerAndActing(body.Entity) {
			continue
		}
		if pos.Dist2(&body.Pos.Now) < pc.Range*pc.Range {
			pc.fire(body, nil)
		}
	}
}

func (pc *ProximityController) proximityOnSector(sector *core.Sector) {
	pc.flags |= behaviors.ProximityOnSector
	for _, pvs := range sector.PVS {
		if pc.ActsOnSectors && sector.Center.Dist2(&pvs.Center) < pc.Range*pc.Range {
			pc.flags |= behaviors.ProximityTargetsSector
			pc.flags &= ^behaviors.ProximityTargetsBody
			pc.fire(nil, sector)
		}
		pc.flags |= behaviors.ProximityTargetsBody
		pc.flags &= ^behaviors.ProximityTargetsSector
		pc.sectorBodies(pvs, &sector.Center)
	}
}

func (pc *ProximityController) proximityOnBody(body *core.Body) {
	// TODO: We should consider the case when the "on" entity is the player.
	//if !pc.isEntityPlayerAndActing(body.Entity) {
	//	return
	//}
	pc.flags &= ^behaviors.ProximityOnSector
	pc.flags |= behaviors.ProximityOnBody
	container := body.Sector()
	for _, sector := range container.PVS {
		if pc.ActsOnSectors && sector.Center.Dist2(&body.Pos.Now) < pc.Range*pc.Range {
			pc.flags |= behaviors.ProximityTargetsSector
			pc.flags &= ^behaviors.ProximityTargetsBody
			pc.fire(nil, sector)
		}
		pc.flags |= behaviors.ProximityTargetsBody
		pc.flags &= ^behaviors.ProximityTargetsSector
		pc.sectorBodies(sector, &body.Pos.Now)
	}
}

func (pc *ProximityController) proximity(proximityEntity ecs.Entity) {
	// TODO: Add InternalSegments
	if sector := core.GetSector(pc.ECS, proximityEntity); sector != nil {
		pc.proximityOnSector(sector)
	} else if body := core.GetBody(pc.ECS, proximityEntity); body != nil && body.SectorEntity != 0 {
		pc.proximityOnBody(body)
	}
}

func (pc *ProximityController) Always() {
	pc.Firing = false
	/*
		We have several factors to consider:
		1. Is the component referring to an entity, attached to one, or both?
		2. What kind of entity is the proximity on? (sector, body, etc...)
		3. What kind of target does this component respond to (sector, body,
		   etc...)

	*/

	// Is the target itself a body or sector?
	pc.flags = behaviors.ProximitySelf
	pc.onEntity = pc.Entity
	pc.proximity(pc.Entity)

	for entity := range pc.Entities {
		pc.flags = behaviors.ProximityRefers
		pc.onEntity = entity
		pc.proximity(entity)
	}
}
