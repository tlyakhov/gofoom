// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"math/rand"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type timedState int

const (
	timedFired timedState = iota
	timedDelayed
	timedReady
)

type ActionController struct {
	ecs.BaseController
	*behaviors.Actor
	State  *behaviors.ActorState
	Body   *core.Body
	Mobile *core.Mobile
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &ActionController{} }, 100)
}

func (ac *ActionController) ComponentID() ecs.ComponentID {
	return behaviors.ActorCID
}

func (ac *ActionController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways | ecs.ControllerRecalculate
}

func (ac *ActionController) Target(target ecs.Attachable) bool {
	if target.GetECS().Simulation.EditorPaused {
		return false
	}
	ac.Actor = target.(*behaviors.Actor)
	ac.State = behaviors.GetActorState(ac.ECS, ac.Entity)
	ac.Body = core.GetBody(ac.ECS, ac.Entity)
	ac.Mobile = core.GetMobile(ac.ECS, ac.Entity)
	return ac.Actor.IsActive() &&
		ac.Body != nil && ac.Body.IsActive() &&
		(ac.State == nil || ac.State.IsActive())
}

func (ac *ActionController) Recalculate() {
	if ac.State == nil {
		ac.State = ac.ECS.NewAttachedComponent(ac.Entity, behaviors.ActorStateCID).(*behaviors.ActorState)
	}

	if ac.State.Action != 0 {
		return
	}

	var d2 float64
	var closest *behaviors.ActionWaypoint

	closestDist2 := math.MaxFloat64

	behaviors.IterateActions(ac.ECS, ac.Start, func(action ecs.Entity, _ *concepts.Vector3) {
		waypoint := behaviors.GetActionWaypoint(ac.ECS, action)
		if waypoint == nil {
			return
		}
		if ac.NoZ {
			d2 = waypoint.P.To2D().Dist2(ac.Body.Pos.Now.To2D())
		} else {
			d2 = waypoint.P.Dist2(&ac.Body.Pos.Now)
		}
		if d2 < closestDist2 {
			closestDist2 = d2
			closest = waypoint
		}
	})

	if closest != nil {
		ac.State.Action = closest.Entity
		ac.State.LastTransition = ac.ECS.Timestamp
	}
}

func (ac *ActionController) timedAction(timed *behaviors.ActionTimed) timedState {
	if timed.Fired.Contains(ac.Body.Entity) {
		return timedFired
	}
	if ac.ECS.Timestamp-ac.State.LastTransition < int64(timed.Delay.Now) {
		return timedDelayed
	}

	timed.Fired.Add(ac.Body.Entity)
	return timedReady
}

func (ac *ActionController) Jump(jump *behaviors.ActionJump) bool {
	if ac.Mobile == nil {
		return true
	}
	state := ac.timedAction(&jump.ActionTimed)

	switch state {
	case timedFired:
		return true
	case timedDelayed:
		return false
	default:
		// TODO: Parameterize this
		ac.Mobile.Force[2] += constants.PlayerJumpForce * 0.5
		return true
	}
}

func (ac *ActionController) Fire(fire *behaviors.ActionFire) bool {
	weapon := behaviors.GetWeaponInstant(ac.ECS, ac.Body.Entity)

	if weapon == nil {
		return true
	}

	state := ac.timedAction(&fire.ActionTimed)

	switch state {
	case timedFired:
		return true
	case timedDelayed:
		return false
	default:
		weapon.FireNextFrame = true
		return true
	}
}

func (ac *ActionController) Waypoint(waypoint *behaviors.ActionWaypoint) bool {
	state := ac.timedAction(&waypoint.ActionTimed)

	switch state {
	case timedDelayed:
		return false
	}

	pos := &ac.Body.Pos.Now
	if ac.Body.Pos.Procedural {
		pos = &ac.Body.Pos.Input
	}

	// Have we reached the target?
	if pos.To2D().Dist2(waypoint.P.To2D()) < 1 {
		return true
	}

	force := waypoint.P.Sub(pos)
	if ac.NoZ {
		force[2] = 0
		if ac.Body.Pos.Procedural {
			ac.Body.Pos.Input[2] = ac.Body.Pos.Now[2]
		}
	}
	dist := force.Length()
	if dist > 0 {
		angle := math.Atan2(force[1], force[0]) * concepts.Rad2deg
		if ac.FaceNextWaypoint && ac.Body.Angle.Procedural {
			ac.Body.Angle.Input = angle
		} else if ac.FaceNextWaypoint {
			ac.Body.Angle.Now = angle
		}
		speed := ac.Speed / dist
		force.MulSelf(speed)
	}
	switch {
	case ac.Mobile != nil:
		// TODO: Is this a hack?
		if !ac.Body.OnGround {
			force.MulSelf(0.1)
		}
		ac.Mobile.Force.AddSelf(force)
	case ac.Body.Pos.Procedural:
		ac.Body.Pos.Input.AddSelf(force.MulSelf(constants.TimeStepS))
		var bc BodyController
		bc.Target(ac.Body)
		bc.findBodySector()
	default:
		ac.Body.Pos.Now.AddSelf(force.MulSelf(constants.TimeStepS))
		var bc BodyController
		bc.Target(ac.Body)
		bc.findBodySector()
	}

	return false
}

func (ac *ActionController) Always() {
	if ac.State == nil {
		ac.Recalculate()
	}

	if ac.State.Action == 0 {
		ac.State.Action = ac.Start
	}

	doTransition := true

	var timed *behaviors.ActionTimed

	if waypoint := behaviors.GetActionWaypoint(ac.ECS, ac.State.Action); waypoint != nil {
		doTransition = ac.Waypoint(waypoint) && doTransition
		timed = &waypoint.ActionTimed
	}

	if jump := behaviors.GetActionJump(ac.ECS, ac.State.Action); jump != nil {
		// Order of ops matters - the && short circuits on false
		doTransition = ac.Jump(jump) && doTransition
		timed = &jump.ActionTimed
	}

	if fire := behaviors.GetActionFire(ac.ECS, ac.State.Action); fire != nil {
		// Order of ops matters - the && short circuits on false
		doTransition = ac.Fire(fire) && doTransition
		timed = &fire.ActionTimed
	}

	if !doTransition {
		return
	}

	if timed != nil {
		timed.Fired.Delete(ac.Body.Entity)
	}

	if t := behaviors.GetActionTransition(ac.ECS, ac.State.Action); t != nil {
		ac.State.LastTransition = ac.ECS.Timestamp
		if len(t.Next) > 0 {
			i := rand.Intn(len(t.Next))
			for next := range t.Next {
				ac.State.Action = next
				if i <= 0 {
					break
				}
				i--
			}
		} else {
			switch ac.Lifetime {
			case dynamic.AnimationLifetimeLoop:
				ac.State.Action = ac.Start
			case dynamic.AnimationLifetimeOnce:
				ac.State.Active = false
			case dynamic.AnimationLifetimeBounce:
			case dynamic.AnimationLifetimeBounceOnce:
			}
		}
	}
}
