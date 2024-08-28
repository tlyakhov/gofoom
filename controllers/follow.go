// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
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
	Body *core.Body
	Path *core.Path
}

func init() {
	ecs.Types().RegisterController(&FollowController{}, 100)
}

func (fc *FollowController) ComponentIndex() int {
	return behaviors.FollowerComponentIndex
}

func (fc *FollowController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways | ecs.ControllerRecalculate
}

func (fc *FollowController) Target(target ecs.Attachable) bool {
	fc.Follower = target.(*behaviors.Follower)
	fc.Body = core.GetBody(fc.Follower.ECS, fc.Follower.Entity)
	fc.Path = core.GetPath(fc.Follower.ECS, fc.Follower.Path)
	return fc.Follower.IsActive() && fc.Body.IsActive() && fc.Path != nil && fc.Path.IsActive()
}

func (fc *FollowController) Recalculate() {
	var d2 float64
	closest := 0
	closestDist2 := math.MaxFloat64
	for i, seg := range fc.Path.Segments {
		if fc.NoZ {
			d2 = seg.P.To2D().Dist2(fc.Body.Pos.Now.To2D())
		} else {
			d2 = seg.P.Dist2(&fc.Body.Pos.Now)
		}
		if d2 < closestDist2 {
			closestDist2 = d2
			closest = i
		}
	}
	fc.Index = closest
}

func (fc *FollowController) Always() {
	target := fc.Path.Segments[fc.Index]

	pos := &fc.Body.Pos.Now
	if fc.Body.Pos.Procedural {
		pos = &fc.Body.Pos.Input
	}

	// Have we reached the target?
	if pos.To2D().Dist2(target.P.To2D()) < 1 {
		end := len(fc.Path.Segments)
		fc.Follower.Index = fc.Follower.Index + 1
		if fc.Follower.Index >= end {
			switch fc.Lifetime {
			case dynamic.AnimationLifetimeLoop:
				fc.Follower.Index = 0
			case dynamic.AnimationLifetimeOnce:
				fc.Follower.Index = end - 1
				fc.Follower.Active = false
			case dynamic.AnimationLifetimeBounce:
				fc.Follower.Index = concepts.Max(end-2, 0)
			case dynamic.AnimationLifetimeBounceOnce:
				fc.Follower.Index = concepts.Max(end-2, 0)
			}
		}
		return
	}

	v := target.P.Sub(pos)
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
