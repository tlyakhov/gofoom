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
	return ecs.ControllerFrame | ecs.ControllerRecalculate
}

func (ac *ActionController) Target(target ecs.Component, e ecs.Entity) bool {
	if ecs.Simulation.EditorPaused {
		return false
	}
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

func (ac *ActionController) Recalculate() {
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

func (ac *ActionController) Waypoint(waypoint *behaviors.ActionWaypoint) bool {
	state := ac.timedAction(&waypoint.ActionTimed)

	switch state {
	case timedDelayed:
		return false
	}

	pos := ac.pos(true)

	// Have we reached the target?
	if pos.To2D().DistSq(waypoint.P.To2D()) < ac.Speed*constants.TimeStepS {
		return true
	}

	force := waypoint.P.Sub(pos)
	if ac.NoZ || ac.Sector != nil {
		force[2] = 0
		if ac.Body != nil && ac.Body.Pos.Procedural {
			ac.Body.Pos.Input[2] = ac.Body.Pos.Now[2]
		}
	}
	dist := force.Length()
	if dist > 0 {
		angle := math.Atan2(force[1], force[0]) * concepts.Rad2deg
		speed := ac.Speed / dist
		force.MulSelf(speed)
		if ac.FaceNextWaypoint {
			switch {
			case ac.Sector != nil && ac.Sector.Transform.Procedural:
				ac.Sector.Transform.Input.SetRotation(angle)
			case ac.Sector != nil:
				ac.Sector.Transform.Now.SetRotation(angle)
			case ac.Body.Angle.Procedural:
				ac.Body.Angle.Input = angle
			default:
				ac.Body.Angle.Now = angle
			}
		}
	}
	switch {
	case ac.Sector != nil && ac.Sector.Transform.Procedural:
		ac.Sector.Transform.Input.TranslateSelf(force.To2D().MulSelf(constants.TimeStepS))
	case ac.Sector != nil:
		ac.Sector.Transform.Now.TranslateSelf(force.To2D().MulSelf(constants.TimeStepS))
	case ac.Mobile != nil:
		// TODO: Is this a hack?
		if !ac.Body.OnGround {
			force.MulSelf(0.1)
		}
		ac.Mobile.Force.AddSelf(force)
	case ac.Body.Pos.Procedural:
		ac.Body.Pos.Input.AddSelf(force.MulSelf(constants.TimeStepS))
		var bc BodyController
		bc.Target(ac.Body, ac.Entity)
		bc.findBodySector()
	default:
		ac.Body.Pos.Now.AddSelf(force.MulSelf(constants.TimeStepS))
		var bc BodyController
		bc.Target(ac.Body, ac.Entity)
		bc.findBodySector()
	}

	return false
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
		ac.Recalculate()
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
