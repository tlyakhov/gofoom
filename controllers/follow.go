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
	Body   *core.Body
	Start  *core.ActionWaypoint
	Action *core.ActionWaypoint
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
	fc.Start = core.GetActionWaypoint(fc.Follower.ECS, fc.Follower.Start)
	fc.Action = core.GetActionWaypoint(fc.Follower.ECS, fc.Follower.Action)
	return fc.Follower.IsActive() && fc.Body.IsActive()
}

func (fc *FollowController) Recalculate() {
	var d2 float64
	var closest *core.ActionWaypoint

	visited := make(containers.Set[ecs.Entity])

	closestDist2 := math.MaxFloat64
	action := fc.Start
	for action != nil {
		if visited.Contains(action.Entity) {
			break
		}
		visited.Add(action.Entity)
		if fc.NoZ {
			d2 = action.P.To2D().Dist2(fc.Body.Pos.Now.To2D())
		} else {
			d2 = action.P.Dist2(&fc.Body.Pos.Now)
		}
		if d2 < closestDist2 {
			closestDist2 = d2
			closest = action
		}
		action = core.GetActionWaypoint(action.ECS, action.Next)
	}
	fc.Action = closest
	fc.Follower.Action = closest.Entity
}

func (fc *FollowController) Always() {
	if fc.Follower.Action == 0 {
		fc.Follower.Action = fc.Follower.Start
		fc.Action = fc.Start
	}

	pos := &fc.Body.Pos.Now
	if fc.Body.Pos.Procedural {
		pos = &fc.Body.Pos.Input
	}

	// Have we reached the target?
	if pos.To2D().Dist2(fc.Action.P.To2D()) < 1 {
		if fc.Action.Next != 0 {
			fc.Follower.Action = fc.Action.Next
		} else {
			switch fc.Lifetime {
			case dynamic.AnimationLifetimeLoop:
				fc.Follower.Action = fc.Follower.Start
			case dynamic.AnimationLifetimeOnce:
				fc.Follower.Active = false
			case dynamic.AnimationLifetimeBounce:
			case dynamic.AnimationLifetimeBounceOnce:
			}
		}
		return
	}

	v := fc.Action.P.Sub(pos)
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
}
