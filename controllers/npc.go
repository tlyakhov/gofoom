// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"math/rand/v2"
	"tlyakhov/gofoom/components/audio"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
)

type NpcController struct {
	ecs.BaseController
	*character.Npc
	Alive *behaviors.Alive
	Body  *core.Body
}

var npcFuncs = [character.NpcStateCount]func(*NpcController){}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &NpcController{} }, 100)
	npcFuncs[character.NpcStateIdle] = npcIdle
	npcFuncs[character.NpcStateSawTarget] = npcOther
	npcFuncs[character.NpcStatePursuit] = npcOther
	npcFuncs[character.NpcStateLostTarget] = npcOther
	npcFuncs[character.NpcStateSearching] = npcOther
	npcFuncs[character.NpcStateDead] = npcOther
}

func (npc *NpcController) ComponentID() ecs.ComponentID {
	return character.NpcCID
}

func (npc *NpcController) Methods() ecs.ControllerMethod {
	return ecs.ControllerFrame
}

func (npc *NpcController) Target(target ecs.Component, e ecs.Entity) bool {
	npc.Entity = e
	npc.Npc = target.(*character.Npc)
	if !npc.Npc.IsActive() {
		return false
	}
	npc.Body = core.GetBody(npc.Entity)
	if npc.Body == nil || !npc.Body.IsActive() {
		return false
	}
	npc.Alive = behaviors.GetAlive(npc.Entity)

	return true
}

func (npc *NpcController) nextState() {
	npc.State = npc.NextState
	switch npc.NextState {
	case character.NpcStateLostTarget:
		//log.Printf("Pursuer %v lost sight of %v!", pc.Entity, enemy.Entity)
		npc.playSound(npc.BarksTargetLost)
		npc.NextState = character.NpcStateSearching
	case character.NpcStateSawTarget:
		//log.Printf("Pursuer %v lost sight of %v!", pc.Entity, enemy.Entity)
		npc.playSound(npc.BarksTargetSeen)
		npc.NextState = character.NpcStatePursuit
	}
}

func (npc *NpcController) processHealthBarks() {
	if npc.Alive.Health.PrevFrame <= npc.Alive.Health.Now {
		return
	}
	switch {
	case npc.Alive.Health.Now > 66:
		npc.playSound(npc.BarksHurtLow)
	case npc.Alive.Health.Now > 33:
		npc.playSound(npc.BarksHurtMed)
	case npc.Alive.Health.Now > 0:
		npc.playSound(npc.BarksHurtMed)
	default:
		npc.playSound(npc.BarksDying)
	}
}

func (npc *NpcController) Frame() {
	if npc.Alive != nil {
		npc.processHealthBarks()
	}
	npcFuncs[npc.State](npc)
	if npc.State != npc.NextState {
		npc.nextState()
	}
}

func npcIdle(npc *NpcController) {
	if ecs.Simulation.SimTimestamp > npc.NextIdleBark {
		if npc.NextIdleBark != 0 {
			npc.playSound(npc.BarksIdle)
		}
		npc.NextIdleBark = ecs.Simulation.SimTimestamp + concepts.MillisToNanos(5000+rand.Float64()*15000)
	}
}

func npcOther(npc *NpcController) {
}

func (npc *NpcController) playSound(sounds ecs.EntityTable) {
	sound := ecs.Entity(0)
	r := rand.IntN(sounds.Len())
	for _, e := range sounds {
		if e == 0 {
			continue
		}
		if r <= 0 {
			sound = e
			break
		}
		r--
	}
	if sound == 0 {
		return
	}

	event, _ := audio.PlaySound(sound, npc.Body.Entity, npc.Body.Entity.String()+" voice", true)
	if event != nil {
		// TODO: Parameterize this
		event.Offset[2] = npc.Body.Size.Now[1] * 0.3
		event.Offset[1] = 0
		event.Offset[0] = 0
	}
}

func NpcMove(body *core.Body, speed float64, delta *concepts.Vector3, face bool) {
	if face {
		angle := math.Atan2(delta[1], delta[0]) * concepts.Rad2deg
		if body.Angle.Procedural {
			body.Angle.Input = angle
		} else {
			body.Angle.Now = angle
		}
	}

	force := &concepts.Vector3{delta[0] * speed, delta[1] * speed, delta[2] * speed}
	mobile := core.GetMobile(body.Entity)
	switch {
	case mobile != nil:
		// TODO: Is this a hack?
		if !body.OnGround {
			force.MulSelf(0.01)
		}
		mobile.Force.AddSelf(force)
	case body.Pos.Procedural:
		body.Pos.Input.AddSelf(force.MulSelf(constants.TimeStepS))
		var bc BodyController
		if bc.Target(body, body.Entity) {
			bc.findBodySector()
		}
	default:
		body.Pos.Now.AddSelf(force.MulSelf(constants.TimeStepS))
		var bc BodyController
		if bc.Target(body, body.Entity) {
			bc.findBodySector()
		}
	}
}
