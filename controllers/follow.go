// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"log"
	"math"
	"math/rand"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type FollowController struct {
	ecs.BaseController
	*behaviors.Follower
	Body   *core.Body
	Mobile *core.Mobile
}

func init() {
	ecs.Types().RegisterController(&FollowController{}, 100)
}

func (fc *FollowController) ComponentID() ecs.ComponentID {
	return behaviors.FollowerCID
}

func (fc *FollowController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways | ecs.ControllerRecalculate
}

func (fc *FollowController) Target(target ecs.Attachable) bool {
	fc.Follower = target.(*behaviors.Follower)
	fc.Body = core.GetBody(fc.ECS, fc.Entity)
	fc.Mobile = core.GetMobile(fc.ECS, fc.Entity)
	return fc.Follower.IsActive() && fc.Body.IsActive()
}

func (fc *FollowController) Recalculate() {
	if fc.Action != 0 {
		return
	}

	var d2 float64
	var closest *behaviors.ActionWaypoint

	closestDist2 := math.MaxFloat64

	behaviors.IterateActions(fc.ECS, fc.Start, func(action ecs.Entity, _ *concepts.Vector3) {
		waypoint := behaviors.GetActionWaypoint(fc.ECS, action)
		if waypoint == nil {
			return
		}
		if fc.NoZ {
			d2 = waypoint.P.To2D().Dist2(fc.Body.Pos.Now.To2D())
		} else {
			d2 = waypoint.P.Dist2(&fc.Body.Pos.Now)
		}
		if d2 < closestDist2 {
			closestDist2 = d2
			closest = waypoint
		}
	})

	if closest != nil {
		fc.Action = closest.Entity
		fc.LastTransition = fc.ECS.Timestamp
	}
}

func (fc *FollowController) Jump(jump *behaviors.ActionJump) bool {
	if fc.Mobile == nil || jump.Fired.Contains(fc.Body.Entity) {
		return true
	}
	if fc.ECS.Timestamp-fc.LastTransition < int64(jump.Delay.Now) {
		return false
	}

	jump.Fired.Add(fc.Body.Entity)

	// TODO: Parameterize this
	fc.Mobile.Force[2] += constants.PlayerJumpForce * 0.5

	return true
}

func (fc *FollowController) Fire(fire *behaviors.ActionFire) bool {
	weapon := behaviors.GetWeaponInstant(fc.ECS, fc.Body.Entity)

	if weapon == nil || fire.Fired.Contains(fc.Body.Entity) {
		return true
	}
	log.Printf("%v < %v?", fc.ECS.Timestamp-fc.LastTransition, int64(fire.Delay.Now))
	if fc.ECS.Timestamp-fc.LastTransition < int64(fire.Delay.Now) {
		return false
	}

	fire.Fired.Add(fc.Body.Entity)

	weapon.FireNextFrame = true

	return true
}

func (fc *FollowController) Waypoint(waypoint *behaviors.ActionWaypoint) bool {
	pos := &fc.Body.Pos.Now
	if fc.Body.Pos.Procedural {
		pos = &fc.Body.Pos.Input
	}

	// Have we reached the target?
	if pos.To2D().Dist2(waypoint.P.To2D()) < 1 {
		return true
	}

	v := waypoint.P.Sub(pos)
	if fc.NoZ {
		v[2] = 0
		if fc.Body.Pos.Procedural {
			fc.Body.Pos.Input[2] = fc.Body.Pos.Now[2]
		}
	}
	dist := v.Length()
	if dist > 0 {
		fc.Body.Angle.Input = math.Atan2(v[1], v[0]) * concepts.Rad2deg
		speed := fc.Speed * constants.TimeStepS / dist
		if speed < 1 {
			v.MulSelf(speed)
		}
	}
	if fc.Body.Pos.Procedural {
		fc.Body.Pos.Input.AddSelf(v)
	} else if fc.Mobile != nil {
		fc.Mobile.Vel.Now = *v
	}

	return false
}

func (fc *FollowController) Always() {
	if fc.Action == 0 {
		fc.Action = fc.Start
	}

	doTransition := true

	var timed *behaviors.ActionTimed

	if waypoint := behaviors.GetActionWaypoint(fc.ECS, fc.Action); waypoint != nil {
		doTransition = fc.Waypoint(waypoint) && doTransition
	}

	if jump := behaviors.GetActionJump(fc.ECS, fc.Action); jump != nil {
		// Order of ops matters - the && short circuits on false
		doTransition = fc.Jump(jump) && doTransition
		timed = &jump.ActionTimed
	}

	if fire := behaviors.GetActionFire(fc.ECS, fc.Action); fire != nil {
		// Order of ops matters - the && short circuits on false
		doTransition = fc.Fire(fire) && doTransition
		timed = &fire.ActionTimed
	}

	if !doTransition {
		return
	}

	if timed != nil {
		timed.Fired.Delete(fc.Body.Entity)
	}

	if t := behaviors.GetActionTransition(fc.ECS, fc.Action); t != nil {
		log.Printf("Transition %v, time from last: %v", fc.Action.Format(fc.ECS), fc.ECS.Timestamp-fc.LastTransition)
		fc.LastTransition = fc.ECS.Timestamp
		if len(t.Next) > 0 {
			i := rand.Intn(len(t.Next))
			for next := range t.Next {
				fc.Action = next
				if i <= 0 {
					break
				}
				i--
			}
		} else {
			switch fc.Lifetime {
			case dynamic.AnimationLifetimeLoop:
				fc.Action = fc.Start
			case dynamic.AnimationLifetimeOnce:
				fc.Active = false
			case dynamic.AnimationLifetimeBounce:
			case dynamic.AnimationLifetimeBounceOnce:
			}
		}
	}
}
