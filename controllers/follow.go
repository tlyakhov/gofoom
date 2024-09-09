// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type FollowController struct {
	ecs.BaseController
	*behaviors.Follower
	Body *core.Body
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
	fc.Body = core.GetBody(fc.Follower.ECS, fc.Follower.Entity)
	return fc.Follower.IsActive() && fc.Body.IsActive()
}

func (fc *FollowController) Recalculate() {
	var d2 float64
	var closest *behaviors.ActionWaypoint

	visited := make(containers.Set[ecs.Entity])

	closestDist2 := math.MaxFloat64
	action := fc.Start
	for action != 0 {
		if visited.Contains(action) {
			break
		}
		visited.Add(action)

		if waypoint := behaviors.GetActionWaypoint(fc.ECS, action); waypoint != nil {
			if fc.NoZ {
				d2 = waypoint.P.To2D().Dist2(fc.Body.Pos.Now.To2D())
			} else {
				d2 = waypoint.P.Dist2(&fc.Body.Pos.Now)
			}
			if d2 < closestDist2 {
				closestDist2 = d2
				closest = waypoint
			}
		}

		if t := behaviors.GetActionTransition(fc.ECS, action); t != nil {
			action = t.Next
		} else {
			action = 0
		}
	}
	fc.Action = closest.Entity
}

func (fc *FollowController) Jump(jump *behaviors.ActionJump) bool {
	if jump.Fired.Contains(fc.Body.Entity) {
		return true
	}
	jump.Fired.Add(fc.Body.Entity)

	fc.Body.Force[2] += constants.PlayerJumpForce

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
	} else {
		fc.Body.Vel.Now = *v
	}

	return false
}

func (fc *FollowController) Always() {
	if fc.Action == 0 {
		fc.Action = fc.Start
	}

	doTransition := true

	if waypoint := behaviors.GetActionWaypoint(fc.ECS, fc.Action); waypoint != nil {
		doTransition = doTransition && fc.Waypoint(waypoint)
	}

	if jump := behaviors.GetActionJump(fc.ECS, fc.Action); jump != nil {
		doTransition = doTransition && fc.Jump(jump)
	}

	if !doTransition {
		return
	}

	if jump := behaviors.GetActionJump(fc.ECS, fc.Action); jump != nil {
		jump.Fired.Delete(fc.Body.Entity)
	}

	if t := behaviors.GetActionTransition(fc.ECS, fc.Action); t != nil {
		if t.Next != 0 {
			fc.Action = t.Next
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
