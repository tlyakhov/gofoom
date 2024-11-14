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
	flags behaviors.ProximityFlags
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
	if player := behaviors.GetPlayer(pc.ECS, entity); player != nil &&
		player.ActionPressed && player.SelectedTarget == 0 {
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
	if !pc.InRange.IsEmpty() {
		pc.InRange.Params = proximityScriptParams
		pc.InRange.Compile()
	}
	if !pc.Enter.IsEmpty() {
		pc.Enter.Params = proximityScriptParams
		pc.Enter.Compile()
	}
	if !pc.Exit.IsEmpty() {
		pc.Exit.Params = proximityScriptParams
		pc.Exit.Compile()
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

func (pc *ProximityController) actScript(body *core.Body, sector *core.Sector, script *core.Script) {
	script.Vars["proximity"] = pc.Proximity
	script.Vars["onEntity"] = pc.Entity
	script.Vars["body"] = body
	script.Vars["sector"] = sector
	script.Vars["flags"] = pc.flags
	script.Act()
}

func (pc *ProximityController) react(target ecs.Entity, body *core.Body, sector *core.Sector) {
	var loaded bool
	var state *behaviors.ProximityState

	key := uint64(uint32(pc.Entity)) | uint64(uint32(target))
	if state, loaded = pc.State.Load(key); !loaded {
		state = pc.ECS.NewAttachedComponent(pc.ECS.NewEntity(), behaviors.ProximityStateCID).(*behaviors.ProximityState)
		state.System = true
		state.Source = pc.Entity
		state.Target = target
		state.PrevStatus = behaviors.ProximityIdle
		state.Status = behaviors.ProximityIdle
		pc.State.Store(key, state)
	}

	if state.PrevStatus != behaviors.ProximityIdle && pc.Hysteresis > 0 && state.LastFired+int64(pc.Hysteresis) > pc.ECS.Timestamp {
		state.Status = behaviors.ProximityWaiting
		return
	}
	state.LastFired = pc.ECS.Timestamp
	if state.PrevStatus == behaviors.ProximityIdle && pc.Enter.IsCompiled() {
		pc.actScript(body, sector, &pc.Enter)
	}
	state.Status = behaviors.ProximityFiring

	if pc.InRange.IsCompiled() {
		pc.actScript(body, sector, &pc.InRange)
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
			pc.react(body.Entity, body, nil)
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
			pc.react(sector.Entity, nil, sector)
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
			pc.react(sector.Entity, nil, sector)
		}
		pc.flags |= behaviors.ProximityTargetsBody
		pc.flags &= ^behaviors.ProximityTargetsSector
		pc.sectorBodies(sector, &body.Pos.Now)
	}
}

func (pc *ProximityController) Always() {
	pc.State.Range(func(key uint64, state *behaviors.ProximityState) bool {
		if state.Source != pc.Entity {
			return true
		}
		state.PrevStatus = state.Status
		state.Status = behaviors.ProximityIdle
		return true
	})

	/*
		We have several factors to consider:
		1. What kind of entity is the proximity on? (sector, body, etc...)
		2. What kind of target does this component respond to (sector, body,
		   etc...)

	*/

	pc.flags = 0
	// TODO: Add InternalSegments
	if sector := core.GetSector(pc.ECS, pc.Entity); sector != nil && sector.Active {
		pc.proximityOnSector(sector)
	} else if body := core.GetBody(pc.ECS, pc.Entity); body != nil && body.SectorEntity != 0 && body.Active {
		pc.proximityOnBody(body)
	}

	pc.State.Range(func(key uint64, state *behaviors.ProximityState) bool {
		if state.Source != pc.Entity {
			return true
		}
		if state.Status == behaviors.ProximityIdle {
			if state.PrevStatus != behaviors.ProximityIdle && pc.Exit.IsCompiled() {
				// TODO: We should save/fill these fields when firing
				pc.actScript(nil, nil, &pc.Exit)
			}
			pc.State.Delete(key)
			pc.ECS.Delete(state.Entity)
		}
		return true
	})
}
