// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"math/rand"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/inventory"
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
	Sector *core.Sector
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &ActionController{} }, 100)
}

func (ac *ActionController) ComponentID() ecs.ComponentID {
	return behaviors.ActorCID
}

func (ac *ActionController) Methods() ecs.ControllerMethod {
	return ecs.ControllerFrame | ecs.ControllerPrecompute
}

func (ac *ActionController) Target(target ecs.Component, e ecs.Entity) bool {
	ac.Entity = e
	ac.Actor = target.(*behaviors.Actor)
	if !ac.Actor.IsActive() {
		return false
	}

	ac.State = behaviors.GetActorState(ac.Entity)

	if ac.State != nil && !ac.State.IsActive() {
		return false
	}

	ac.Body = core.GetBody(ac.Entity)
	ac.Mobile = core.GetMobile(ac.Entity)
	if ac.Body == nil {
		ac.Sector = core.GetSector(ac.Entity)
	}
	return (ac.Body != nil && ac.Body.IsActive()) ||
		(ac.Sector != nil && ac.Sector.IsActive())
}

func (ac *ActionController) pos(checkProcedural bool) *concepts.Vector3 {
	if ac.Body == nil {
		if checkProcedural && ac.Sector.Transform.Procedural {
			return &ac.Sector.Center.Now
		}
		return &ac.Sector.Center.Now
	}

	if checkProcedural && ac.Body.Pos.Procedural {
		return &ac.Body.Pos.Input
	}
	return &ac.Body.Pos.Now
}

func (ac *ActionController) startSector() *core.Sector {
	if ac.Body == nil {
		// TODO: need to consider overlaps
		return ac.Sector
	}
	return ac.Body.Sector()
}

func (ac *ActionController) Precompute() {
	if ac.State == nil {
		ac.State = ecs.NewAttachedComponent(ac.Entity, behaviors.ActorStateCID).(*behaviors.ActorState)
	}

	if ac.State.Action != 0 {
		return
	}

	var d2 float64
	var closest *behaviors.ActionWaypoint

	closestDistSq := math.MaxFloat64

	behaviors.IterateActions(ac.Start, func(action ecs.Entity, _ *concepts.Vector3) {
		waypoint := behaviors.GetActionWaypoint(action)
		if waypoint == nil {
			return
		}
		if ac.NoZ {
			d2 = waypoint.P.To2D().DistSq(ac.pos(false).To2D())
		} else {
			d2 = waypoint.P.DistSq(ac.pos(false))
		}
		if d2 < closestDistSq {
			closestDistSq = d2
			closest = waypoint
		}
	})

	if closest != nil {
		ac.State.Action = closest.Entity
		ac.State.LastTransition = ecs.Simulation.SimTimestamp
	}
}

func (ac *ActionController) timedAction(timed *behaviors.ActionTimed) timedState {
	if timed.Fired.Contains(ac.Entity) {
		return timedFired
	}
	if ecs.Simulation.SimTimestamp-ac.State.LastTransition < concepts.MillisToNanos(timed.Delay.Now) {
		return timedDelayed
	}

	timed.Fired.Add(ac.Entity)
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
	weapon := inventory.GetWeapon(ac.Entity)

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
		weapon.Intent = inventory.WeaponFire
		return true
	}
}

func (ac *ActionController) Face(face *behaviors.ActionFace) bool {
	state := ac.timedAction(&face.ActionTimed)

	switch state {
	case timedDelayed:
		return false
	}

	angle := 0.0

	switch {
	case ac.Sector != nil && ac.Sector.Transform.Procedural:
		angle, _, _ = ac.Sector.Transform.Input.GetTransform()
	case ac.Sector != nil:
		angle, _, _ = ac.Sector.Transform.Now.GetTransform()
	case ac.Body.Angle.Procedural:
		angle = ac.Body.Angle.Input
	default:
		angle = ac.Body.Angle.Now
	}

	target := face.Angle
	//	log.Printf("Target: %v, Angle: %v", target, angle)
	concepts.MinimizeAngleDistance(angle, &target)
	//	log.Printf("Minimized Target: %v", target)

	// Have we reached the target?
	if math.Abs(target-angle) < ac.AngularSpeed*constants.TimeStepS {
		return true
	}

	delta := ac.AngularSpeed * constants.TimeStepS
	if target-angle < 0 {
		delta = -delta
	}

	switch {
	case ac.Sector != nil && ac.Sector.Transform.Procedural:
		ac.Sector.Transform.Input.SetRotation(angle + delta)
	case ac.Sector != nil:
		ac.Sector.Transform.Now.SetRotation(angle + delta)
	case ac.Body.Angle.Procedural:
		ac.Body.Angle.Input += delta
	default:
		ac.Body.Angle.Now += delta
	}
	return false
}

func (ac *ActionController) Frame() {
	if ac.State == nil {
		ac.Precompute()
	}

	if ac.State.Action == 0 {
		ac.State.Action = ac.Start
	}

	doTransition := true
	ignoreFace := false

	var timed *behaviors.ActionTimed

	if waypoint := behaviors.GetActionWaypoint(ac.State.Action); waypoint != nil {
		// Order of ops matters - the && short circuits on false
		doTransition = ac.Waypoint(waypoint) && doTransition
		timed = &waypoint.ActionTimed
		ignoreFace = ac.FaceNextWaypoint
	}

	if face := behaviors.GetActionFace(ac.State.Action); face != nil && !ignoreFace {
		// Order of ops matters - the && short circuits on false
		doTransition = ac.Face(face) && doTransition
		timed = &face.ActionTimed
	}

	if jump := behaviors.GetActionJump(ac.State.Action); jump != nil {
		// Order of ops matters - the && short circuits on false
		doTransition = ac.Jump(jump) && doTransition
		timed = &jump.ActionTimed
	}

	if fire := behaviors.GetActionFire(ac.State.Action); fire != nil {
		// Order of ops matters - the && short circuits on false
		doTransition = ac.Fire(fire) && doTransition
		timed = &fire.ActionTimed
	}

	if !doTransition {
		return
	}

	if timed != nil {
		timed.Fired.Delete(ac.Entity)
	}

	if t := behaviors.GetActionTransition(ac.State.Action); t != nil {
		ac.State.LastTransition = ecs.Simulation.SimTimestamp
		if len(t.Next) > 0 {
			i := rand.Intn(len(t.Next))
			for _, next := range t.Next {
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
				ac.State.Flags &= ^ecs.ComponentActive
			case dynamic.AnimationLifetimeBounce:
			case dynamic.AnimationLifetimeBounceOnce:
			}
		}
	}
}
