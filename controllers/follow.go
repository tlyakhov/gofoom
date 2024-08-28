// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
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
	return ecs.ControllerAlways
}

func (fc *FollowController) Target(target ecs.Attachable) bool {
	fc.Follower = target.(*behaviors.Follower)
	fc.Body = core.GetBody(fc.Follower.ECS, fc.Follower.Entity)
	fc.Path = core.GetPath(fc.Follower.ECS, fc.Follower.Path)
	return fc.Follower.IsActive() && fc.Body.IsActive() && fc.Path != nil && fc.Path.IsActive()
}

func (fc *FollowController) Always() {
	start := fc.Path.Segments[fc.Index]
	end := start.Next
	if end == nil {
		// TODO: More options, loop/bounce/etc
		end = fc.Path.Segments[0]
	}

	pos := &fc.Body.Pos.Now
	if fc.Body.Pos.Procedural {
		pos = &fc.Body.Pos.Input
	}

	// Have we reached the target?
	if pos.To2D().Dist2(end.P.To2D()) < 1 {
		fc.Follower.Index = (fc.Follower.Index + 1) % len(fc.Path.Segments)
		return
	}

	v := end.P.Sub(pos)
	if fc.NoZ {
		v[2] = 0
		if fc.Body.Pos.Procedural {
			fc.Body.Pos.Input[2] = fc.Body.Pos.Now[2]
		}
	}
	fc.Body.Angle.Input = math.Atan2(v[1], v[0]) * concepts.Rad2deg
	dist := v.Length()
	if dist > 0 {
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
