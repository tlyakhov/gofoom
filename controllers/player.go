// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"

	"tlyakhov/gofoom/components/audio"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/ecs"

	"tlyakhov/gofoom/constants"
)

type PlayerController struct {
	ecs.BaseController
	*character.Player
	Alive   *behaviors.Alive
	Body    *core.Body
	Mobile  *core.Mobile
	Carrier *inventory.Carrier
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &PlayerController{} }, 100)
}

func (pc *PlayerController) ComponentID() ecs.ComponentID {
	return character.PlayerCID
}

func (pc *PlayerController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways | ecs.ControllerRecalculate
}

func (pc *PlayerController) Target(target ecs.Component, e ecs.Entity) bool {
	pc.Entity = e
	pc.Player = target.(*character.Player)
	if !pc.Player.IsActive() || pc.Player.Spawn {
		return false
	}
	pc.Body = core.GetBody(pc.Entity)
	if pc.Body == nil || !pc.Body.IsActive() {
		return false
	}
	pc.Alive = behaviors.GetAlive(pc.Entity)
	if pc.Alive == nil || !pc.Alive.IsActive() {
		return false
	}
	pc.Carrier = inventory.GetCarrier(pc.Entity)
	if pc.Carrier == nil || !pc.Carrier.IsActive() {
		return false
	}
	pc.Mobile = core.GetMobile(pc.Entity)
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
		pc.Crouching = false
	} else {
		pc.Body.Size.Now[1] = constants.PlayerHeight
	}

	bob := math.Sin(pc.Bob) * 1.5
	pc.CameraZ = pc.Body.Pos.Render[2] + pc.Body.Size.Render[1]*0.5 + bob - 5

	if sector := pc.Body.Sector(); sector != nil {
		fz, cz := sector.ZAt(pc.Body.Pos.Render.To2D())
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

func (pc *PlayerController) Recalculate() {
	pc.bob(pc.Underwater())
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
	for _, e := range pc.Carrier.Slots {
		if e == 0 {
			continue
		}
		slot := inventory.GetSlot(e)
		// TODO: Put this into an Carrier controller
		if slot.Carrier != pc.Carrier {
			slot.Carrier = pc.Carrier
		}

		if slot.Count.Now <= 0 {
			continue
		}
		if w := inventory.GetWeapon(e); w != nil {
			pc.Carrier.SelectedWeapon = e
			break
		}
	}

	// Audio
	mixer := ecs.Singleton(audio.MixerCID).(*audio.Mixer)
	mixer.PollSources()
	mixer.SetListenerPosition(&pc.Body.Pos.Now)
	dy, dx := math.Sincos(pc.Body.Angle.Now)
	mixer.SetListenerOrientation(&concepts.Vector3{dx * constants.UnitsPerMeter, dy * constants.UnitsPerMeter, 0})
	mixer.SetListenerVelocity(&pc.Mobile.Vel.Now)

	// This section handles frobbing

	// Figure out closest body out of the ones the player can select
	prevTarget := pc.SelectedTarget
	pc.SelectedTarget = 0
	closestDistSq := math.MaxFloat64
	for e := range pc.HoveringTargets {
		pt := behaviors.GetPlayerTargetable(e)
		if pt == nil {
			continue
		}
		distSq := pt.Pos(e).DistSq(&pc.Body.Pos.Now)
		if distSq > closestDistSq {
			continue
		}
		closestDistSq = distSq
		pc.SelectedTarget = e
	}
	// If our selection has changed, run scripts
	if pc.SelectedTarget != prevTarget {
		if prevTarget != 0 {
			if pt := behaviors.GetPlayerTargetable(prevTarget); pt != nil && pt.UnSelected.IsCompiled() {
				pt.UnSelected.Vars["body"] = core.GetBody(pc.SelectedTarget)
				pt.UnSelected.Vars["player"] = pc.Player
				pt.UnSelected.Vars["carrier"] = pc.Carrier
				pt.UnSelected.Act()
			}
		}
		if pc.SelectedTarget != 0 {
			if pt := behaviors.GetPlayerTargetable(pc.SelectedTarget); pt != nil && pt.Selected.IsCompiled() {
				pt.Selected.Vars["body"] = core.GetBody(pc.SelectedTarget)
				pt.Selected.Vars["player"] = pc.Player
				pt.Selected.Vars["carrier"] = pc.Carrier
				pt.Selected.Act()
			}
		}
	}
	// If we have a selected item and the player is pressing the key, frob!
	if pc.SelectedTarget != 0 && pc.ActionPressed {
		if pt := behaviors.GetPlayerTargetable(pc.SelectedTarget); pt != nil && pt.Frob.IsCompiled() {
			pt.Frob.Vars["body"] = core.GetBody(pc.SelectedTarget)
			pt.Frob.Vars["player"] = pc.Player
			pt.Frob.Vars["carrier"] = pc.Carrier
			pt.Frob.Act()
		}
	}
	// Reset our potential selection for next frame
	pc.ActionPressed = false
	pc.HoveringTargets = make(containers.Set[ecs.Entity])
}

func MovePlayer(e ecs.Entity, angle float64) {
	if ecs.Simulation.EditorPaused {
		MovePlayerNoClip(e, angle)
	} else {
		MovePlayerForce(e, angle)
	}
}

func MovePlayerForce(e ecs.Entity, angle float64) {
	p := core.GetBody(e)
	m := core.GetMobile(e)
	if p == nil || m == nil {
		return
	}
	uw := behaviors.GetUnderwater(p.SectorEntity) != nil
	dy, dx := math.Sincos(angle * concepts.Deg2rad)
	dy *= constants.PlayerWalkForce
	dx *= constants.PlayerWalkForce

	if uw || p.OnGround {
		m.Force[0] += dx
		m.Force[1] += dy
	}
}

func MovePlayerNoClip(e ecs.Entity, angle float64) {
	p := core.GetBody(e)
	m := core.GetMobile(e)
	if p == nil || m == nil {
		return
	}
	dy, dx := math.Sincos(angle * concepts.Deg2rad)
	dy *= constants.PlayerWalkForce
	dx *= constants.PlayerWalkForce
	player := character.GetPlayer(e)
	p.Pos.Now[0] += dx * 0.02 / m.Mass
	p.Pos.Now[1] += dy * 0.02 / m.Mass
	sector := p.RenderSector()
	if sector != nil {
		p.SectorEntity = sector.Entity
		p.Pos.Now[2] = sector.Center.Now[2]
		player.CameraZ = p.Pos.Now[2]
	}
}
