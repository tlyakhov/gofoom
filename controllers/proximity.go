// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
)

type ProximityController struct {
	ecs.BaseController
	*behaviors.Proximity
	TargetBody   *core.Body
	TargetSector *core.Sector
	flags        behaviors.ProximityFlags

	tree *core.Quadtree
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

func (pc *ProximityController) Target(target ecs.Attachable, e ecs.Entity) bool {
	pc.Entity = e
	pc.Proximity = target.(*behaviors.Proximity)
	return pc.Proximity.IsActive()
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
		if pc.Universe.Component(e, cid) == nil {
			return false
		}
	}
	return true
}

func (pc *ProximityController) actScript(script *core.Script) {
	script.Vars["proximity"] = pc.Proximity
	script.Vars["onEntity"] = pc.Entity
	script.Vars["body"] = pc.TargetBody
	script.Vars["sector"] = pc.TargetSector
	script.Vars["flags"] = pc.flags
	script.Act()
}

func (pc *ProximityController) react(target ecs.Entity) {
	var loaded bool
	var state *behaviors.ProximityState

	key := uint64(uint32(pc.Entity)) | uint64(uint32(target))
	if state, loaded = pc.State.Load(key); !loaded {
		// TODO: Think through the lifecycle of these. Should they be serialized?
		state = pc.Universe.NewAttachedComponent(pc.Universe.NewEntity(), behaviors.ProximityStateCID).(*behaviors.ProximityState)
		state.Attached.Flags |= ecs.ComponentInternal
		state.Source = pc.Entity
		state.Target = target
		state.PrevStatus = behaviors.ProximityIdle
		state.Status = behaviors.ProximityIdle
		state.Flags = pc.flags
		pc.State.Store(key, state)
	}

	if state.PrevStatus != behaviors.ProximityIdle && pc.Hysteresis > 0 && state.LastFired+int64(pc.Hysteresis) > pc.Universe.Timestamp {
		state.Status = behaviors.ProximityWaiting
		return
	}
	state.LastFired = pc.Universe.Timestamp
	if state.PrevStatus == behaviors.ProximityIdle && pc.Enter.IsCompiled() {
		pc.actScript(&pc.Enter)
	}
	state.Status = behaviors.ProximityFiring

	if pc.InRange.IsCompiled() {
		pc.actScript(&pc.InRange)
	}

}

func (pc *ProximityController) proximityOnSector(sector *core.Sector) {
	pc.flags |= behaviors.ProximityOnSector

	pc.flags |= behaviors.ProximityTargetsBody
	pc.flags &= ^behaviors.ProximityTargetsSector
	pc.tree.Root.RangeCircle(sector.Center.To2D(), pc.Range, func(b *core.Body) bool {
		if !b.IsActive() || !pc.isValid(b.Entity) {
			return true
		}
		pc.TargetBody = b
		pc.TargetSector = nil
		pc.react(b.Entity)
		return true
	})

	// TODO: Reimplement sector<->sector proximity
	/*	sector.PVS.Range(func(e uint32) {
		pvs := core.GetSector(pc.Universe, ecs.Entity(e))
		if !pvs.Active || pvs.Entity == pc.Entity {
			return
		}
		if pc.ActsOnSectors &&
			sector.Center.Dist2(&pvs.Center) < pc.Range*pc.Range && pc.isValid(pvs.Entity) {
			pc.flags |= behaviors.ProximityTargetsSector
			pc.flags &= ^behaviors.ProximityTargetsBody
			pc.TargetBody = nil
			pc.TargetSector = sector
			pc.react(sector.Entity)
		}
		pc.flags |= behaviors.ProximityTargetsBody
		pc.flags &= ^behaviors.ProximityTargetsSector
		pc.sectorBodies(pvs, &sector.Center)
	})*/
}

func (pc *ProximityController) proximityOnBody(body *core.Body) {
	pc.flags &= ^behaviors.ProximityOnSector
	pc.flags |= behaviors.ProximityOnBody

	pc.flags |= behaviors.ProximityTargetsBody
	pc.flags &= ^behaviors.ProximityTargetsSector
	pc.tree.Root.RangeCircle(body.Pos.Now.To2D(), pc.Range, func(b *core.Body) bool {
		if !b.IsActive() || b == body || !pc.isValid(b.Entity) {
			return true
		}
		pc.TargetBody = b
		pc.TargetSector = nil
		pc.react(b.Entity)
		return true
	})

	// TODO: Implement proximityOnBody<->sector tests
	/*container.PVS.Range(func(e uint32) {
		sector := core.GetSector(pc.Universe, ecs.Entity(e))
		if sector.Active && pc.ActsOnSectors &&
			sector.Center.Dist2(&body.Pos.Now) < pc.Range*pc.Range &&
			pc.isValid(sector.Entity) {
			pc.flags |= behaviors.ProximityTargetsSector
			pc.flags &= ^behaviors.ProximityTargetsBody
			pc.TargetBody = nil
			pc.TargetSector = sector
			pc.react(sector.Entity)
		}
	})*/
}

func (pc *ProximityController) Always() {
	if pc.tree == nil {
		pc.tree = pc.Universe.Singleton(core.QuadtreeCID).(*core.Quadtree)
	}

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
	if sector := core.GetSector(pc.Universe, pc.Entity); sector != nil && sector.IsActive() {
		pc.proximityOnSector(sector)
	} else if body := core.GetBody(pc.Universe, pc.Entity); body != nil && body.SectorEntity != 0 && body.IsActive() {
		pc.proximityOnBody(body)
	}

	pc.State.Range(func(key uint64, state *behaviors.ProximityState) bool {
		if state.Source != pc.Entity {
			return true
		}
		if state.Status == behaviors.ProximityIdle {
			if state.PrevStatus != behaviors.ProximityIdle && pc.Exit.IsCompiled() {
				if state.Flags&behaviors.ProximityOnBody != 0 {
					pc.TargetBody = core.GetBody(pc.Universe, state.Target)
					pc.TargetSector = nil
				} else if state.Flags&behaviors.ProximityOnSector != 0 {
					pc.TargetSector = core.GetSector(pc.Universe, state.Target)
					pc.TargetBody = nil
				}
				pc.actScript(&pc.Exit)
			}
			pc.State.Delete(key)
			pc.Universe.Delete(state.Entity)
		}
		return true
	})
}
