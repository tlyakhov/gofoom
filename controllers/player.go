// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"strconv"

	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"tlyakhov/gofoom/constants"
)

type PlayerController struct {
	ecs.BaseController
	*behaviors.Player
	Alive  *behaviors.Alive
	Body   *core.Body
	Mobile *core.Mobile
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &PlayerController{} }, 100)
}

func (pc *PlayerController) ComponentID() ecs.ComponentID {
	return behaviors.PlayerCID
}

func (pc *PlayerController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways
}

func (pc *PlayerController) Target(target ecs.Attachable, e ecs.Entity) bool {
	pc.Entity = e
	pc.Player = target.(*behaviors.Player)
	if !pc.Player.IsActive() || pc.Player.Spawn {
		return false
	}
	pc.Body = core.GetBody(pc.ECS, pc.Entity)
	if pc.Body == nil || !pc.Body.IsActive() {
		return false
	}
	pc.Alive = behaviors.GetAlive(pc.ECS, pc.Entity)
	if pc.Alive == nil || !pc.Alive.IsActive() {
		return false
	}
	pc.Mobile = core.GetMobile(pc.ECS, pc.Entity)
	return pc.Mobile != nil && pc.Mobile.IsActive()
}

func (pc *PlayerController) bob(uw bool) {
	if uw {
		pc.Bob += 0.015
	} else {
		pc.Bob += pc.Mobile.Vel.Now.To2D().Length() / 64.0
	}

	for pc.Bob > math.Pi*2 {
		pc.Bob -= math.Pi * 2
	}

	// TODO: There's a bug here: this can cause a player<->floor collision that
	// has to be resolved by shoving the player upwards, making an uncrouch into
	// an unintentional jump.
	if pc.Crouching {
		pc.Body.Size.Now[1] = constants.PlayerCrouchHeight
	} else {
		pc.Body.Size.Now[1] = constants.PlayerHeight
	}

	bob := math.Sin(pc.Bob) * 1.5
	pc.CameraZ = pc.Body.Pos.Render[2] + pc.Body.Size.Render[1]*0.5 + bob - 5

	if sector := pc.Body.Sector(); sector != nil {
		fz, cz := sector.ZAt(dynamic.DynamicRender, pc.Body.Pos.Render.To2D())
		fz += constants.IntersectEpsilon
		cz -= constants.IntersectEpsilon
		if pc.CameraZ < fz {
			pc.CameraZ = fz
		}
		if pc.CameraZ > cz {
			pc.CameraZ = cz
		}
	}
}

func (pc *PlayerController) Always() {
	uw := pc.Underwater()
	if uw {
		pc.FrameTint = concepts.Vector4{0.29, 0.58, 1, 0.35}
	} else {
		pc.FrameTint = concepts.Vector4{}
	}
	pc.Alive.Tint(&pc.FrameTint)

	pc.bob(uw)

	// If we have a weapon, select it
	// TODO: This should be handled by an inventory UI of some kind,
	// or at least a 1...N quick-select
	/*for _, item := range pc.Inventory {
		if w := behaviors.GetWeaponInstant(pc.ECS, item.Entity); w != nil {
			pc.CurrentWeapon = item.Entity
		}
	}*/

	// This section handles frobbing

	// Figure out closest body out of the ones the player can select
	prevTarget := pc.SelectedTarget
	pc.SelectedTarget = 0
	closestDist2 := math.MaxFloat64
	for e := range pc.HoveringTargets {
		body := core.GetBody(pc.ECS, e)
		if body == nil {
			continue
		}
		d2 := body.Pos.Now.Dist2(&pc.Body.Pos.Now)
		if d2 > closestDist2 {
			continue
		}
		closestDist2 = d2
		pc.SelectedTarget = e
	}
	// If our selection has changed, run scripts
	if pc.SelectedTarget != prevTarget {
		if prevTarget != 0 {
			if pt := behaviors.GetPlayerTargetable(pc.ECS, prevTarget); pt != nil && pt.UnSelected.IsCompiled() {
				pt.UnSelected.Vars["body"] = core.GetBody(pc.ECS, pc.SelectedTarget)
				pt.UnSelected.Vars["player"] = pc.Player
				pt.UnSelected.Act()
			}
		}
		if pc.SelectedTarget != 0 {
			if pt := behaviors.GetPlayerTargetable(pc.ECS, pc.SelectedTarget); pt != nil && pt.Selected.IsCompiled() {
				pt.Selected.Vars["body"] = core.GetBody(pc.ECS, pc.SelectedTarget)
				pt.Selected.Vars["player"] = pc.Player
				pt.Selected.Act()
			}
		}
	}
	// If we have a selected item and the player is pressing the key, frob!
	if pc.SelectedTarget != 0 && pc.ActionPressed {
		if pt := behaviors.GetPlayerTargetable(pc.ECS, pc.SelectedTarget); pt != nil && pt.Frob.IsCompiled() {
			pt.Frob.Vars["body"] = core.GetBody(pc.ECS, pc.SelectedTarget)
			pt.Frob.Vars["player"] = pc.Player
			pt.Frob.Act()
		}
	}
	// Reset our potential selection for next frame
	pc.HoveringTargets = make(containers.Set[ecs.Entity])
}

func MovePlayer(db *ecs.ECS, e ecs.Entity, angle float64) {
	if db.EditorPaused {
		MovePlayerNoClip(db, e, angle)
	} else {
		MovePlayerForce(db, e, angle)
	}
}

func MovePlayerForce(db *ecs.ECS, e ecs.Entity, angle float64) {
	p := core.GetBody(db, e)
	m := core.GetMobile(db, e)
	if p == nil || m == nil {
		return
	}
	uw := behaviors.GetUnderwater(db, p.SectorEntity) != nil
	dy, dx := math.Sincos(angle * concepts.Deg2rad)
	dy *= constants.PlayerWalkForce
	dx *= constants.PlayerWalkForce

	if uw || p.OnGround {
		m.Force[0] += dx
		m.Force[1] += dy
	}
}

func MovePlayerNoClip(db *ecs.ECS, e ecs.Entity, angle float64) {
	p := core.GetBody(db, e)
	m := core.GetMobile(db, e)
	if p == nil || m == nil {
		return
	}
	dy, dx := math.Sincos(angle * concepts.Deg2rad)
	dy *= constants.PlayerWalkForce
	dx *= constants.PlayerWalkForce
	player := behaviors.GetPlayer(db, e)
	p.Pos.Now[0] += dx * 0.02 / m.Mass
	p.Pos.Now[1] += dy * 0.02 / m.Mass
	sector := p.RenderSector()
	if sector != nil {
		p.SectorEntity = sector.Entity
		p.Pos.Now[2] = sector.Center[2]
		player.CameraZ = p.Pos.Now[2]
	}
}

func PickUpInventoryItem(p *behaviors.Player, item *behaviors.InventoryItem) {
	for _, slot := range p.Inventory {
		if !slot.ValidClasses.Contains(item.Class) {
			continue
		}
		if slot.Count.Now >= slot.Limit {
			p.Notices.Push("Can't pick up more " + item.Class)
			return
		}
		toAdd := concepts.Min(item.Count.Now, slot.Limit-slot.Count.Now)
		slot.Count.Now += toAdd
		p.Notices.Push("Picked up " + strconv.Itoa(toAdd) + " " + item.Class)
		item.Count.Now -= toAdd
		// Disable all the entity components
		for _, c := range item.ECS.AllComponents(item.Entity) {
			if c != nil {
				c.Base().Active = false
			}
		}
	}
}
