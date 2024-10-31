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
	ecs.Types().RegisterController(func() ecs.Controller { return &ProximityController{} }, 100)
}

func (pc *ProximityController) ComponentID() ecs.ComponentID {
	return behaviors.ProximityCID
}

func (pc *ProximityController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways | ecs.ControllerRecalculate
}

func (pc *ProximityController) Target(target ecs.Attachable) bool {
	pc.Proximity = target.(*behaviors.Proximity)
	return pc.Proximity.IsActive()
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

var proximityScriptParams = []core.ScriptParam{
	{Name: "proximity", TypeName: "*behaviors.Proximity"},
	{Name: "onEntity", TypeName: "ecs.Entity"},
	{Name: "body", TypeName: "*core.Body"},
	{Name: "sector", TypeName: "*core.Sector"},
	{Name: "flags", TypeName: "behaviors.ProximityFlags"},
}

func (pc *ProximityController) Recalculate() {
	for _, script := range pc.Scripts {
		script.Params = proximityScriptParams
		script.Compile()
	}
}

func (pc *ProximityController) isValid(e ecs.Entity) bool {
	for cid := range pc.ValidComponents {
		if pc.ECS.Component(e, cid) == nil {
			return false
		}
	}
	return true
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
		if !body.Active || body.Entity == pc.Entity {
			continue
		}
		if (pc.flags&behaviors.ProximityTargetsBody) != 0 &&
			!pc.isEntityPlayerAndActing(body.Entity) {
			continue
		}
		if pos.Dist2(&body.Pos.Now) < pc.Range*pc.Range && pc.isValid(body.Entity) {
			pc.fire(body, nil)
		}
	}
}

func (pc *ProximityController) proximityOnSector(sector *core.Sector) {
	pc.flags |= behaviors.ProximityOnSector
	for _, pvs := range sector.PVS {
		if !pvs.Active || pvs.Entity == pc.Entity {
			continue
		}
		if pc.ActsOnSectors &&
			sector.Center.Dist2(&pvs.Center) < pc.Range*pc.Range && pc.isValid(pvs.Entity) {
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
		if sector.Active && pc.ActsOnSectors &&
			sector.Center.Dist2(&body.Pos.Now) < pc.Range*pc.Range &&
			pc.isValid(sector.Entity) {
			pc.flags |= behaviors.ProximityTargetsSector
			pc.flags &= ^behaviors.ProximityTargetsBody
			pc.fire(nil, sector)
		}
		pc.flags |= behaviors.ProximityTargetsBody
		pc.flags &= ^behaviors.ProximityTargetsSector
		pc.sectorBodies(sector, &body.Pos.Now)
	}
}

func (pc *ProximityController) Always() {
	pc.Firing = false
	/*
		We have several factors to consider:
		1. What kind of entity is the proximity on? (sector, body, etc...)
		2. What kind of target does this component respond to (sector, body,
		   etc...)

	*/

	pc.flags = 0
	pc.onEntity = pc.Entity
	// TODO: Add InternalSegments
	if sector := core.GetSector(pc.ECS, pc.Entity); sector != nil && sector.Active {
		pc.proximityOnSector(sector)
	} else if body := core.GetBody(pc.ECS, pc.Entity); body != nil && body.SectorEntity != 0 && body.Active {
		pc.proximityOnBody(body)
	}
}
