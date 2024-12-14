// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"math/rand"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
)

type ParticleController struct {
	ecs.BaseController
	*behaviors.ParticleEmitter
	Body   *core.Body
	Mobile *core.Mobile
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &ParticleController{} }, 100)
}

func (pc *ParticleController) ComponentID() ecs.ComponentID {
	return behaviors.ParticleEmitterCID
}

func (pc *ParticleController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways
}

func (pc *ParticleController) Target(target ecs.Attachable, e ecs.Entity) bool {
	pc.Entity = e
	pc.ParticleEmitter = target.(*behaviors.ParticleEmitter)
	if !pc.ParticleEmitter.IsActive() {
		return false
	}
	pc.Body = core.GetBody(pc.ECS, pc.Entity)
	if pc.Body == nil || !pc.Body.IsActive() {
		return false
	}
	pc.Mobile = core.GetMobile(pc.ECS, pc.Entity)
	return true
}

func (pc *ParticleController) Always() {
	toRemove := make([]ecs.Entity, 0, 10)
	for e, timestamp := range pc.Spawned {
		age := float64(pc.ECS.Timestamp - timestamp)

		if vis := materials.GetVisible(pc.ECS, e); vis != nil {
			fade := (age - (pc.Lifetime - pc.FadeTime)) / pc.FadeTime
			vis.Opacity = 1.0 - concepts.Clamp(fade, 0.0, 1.0)
		}

		if age < pc.Lifetime {
			continue
		}
		toRemove = append(toRemove, e)
		pc.ECS.Delete(e)
	}

	for _, e := range toRemove {
		delete(pc.Spawned, e)
	}

	if pc.Source == 0 || len(pc.Spawned) > pc.Limit {
		return
	}
	// Our goal is to spawn ~pc.Limit particles across pc.Lifetime.
	// Approximate the probability that we should spawn a particle this frame.
	// For example, for 10 particles over 10,000ms, we would get:
	// 10/(10000/7.8) = 0.0078
	probability := float64(pc.Limit) / (pc.Lifetime / constants.TimeStep)
	// if this probability is > 1, we need to spawn more than one particle per
	// frame.
	iterations := int(probability) + 1
	probability /= float64(iterations)

	for range iterations {
		if rand.Float64() > probability {
			return
		}

		e := pc.ECS.NewEntity()
		pc.Spawned[e] = pc.ECS.Timestamp
		body := pc.ECS.NewAttachedComponent(e, core.BodyCID).(*core.Body)
		body.System = true
		vis := pc.ECS.NewAttachedComponent(e, materials.VisibleCID).(*materials.Visible)
		vis.System = true
		mobile := pc.ECS.NewAttachedComponent(e, core.MobileCID).(*core.Mobile)
		mobile.System = true

		pc.ECS.Link(e, pc.Source)
		body.Pos.Now.From(&pc.Body.Pos.Now)
		hAngle := pc.Body.Angle.Now + (rand.Float64()-0.5)*pc.XYSpread
		vAngle := (rand.Float64() - 0.5) * pc.ZSpread
		mobile.Vel.Now[0] = math.Cos(hAngle*concepts.Deg2rad) * pc.Vel
		mobile.Vel.Now[1] = math.Sin(hAngle*concepts.Deg2rad) * pc.Vel
		mobile.Vel.Now[2] = math.Sin(vAngle*concepts.Deg2rad) * pc.Vel
		mobile.Mass = 0.25
		mobile.CrBody = core.CollideNone
		mobile.CrPlayer = core.CollideNone
		mobile.CrWall = core.CollideBounce
		var bc BodyController
		bc.Target(body, e)
		bc.Enter(pc.Body.Sector())
	}
}
