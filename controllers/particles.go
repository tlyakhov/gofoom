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
	Body           *core.Body
	Spawner        *behaviors.Spawner
	BodyController BodyController
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &ParticleController{} }, 100)
}

func (pc *ParticleController) ComponentID() ecs.ComponentID {
	return behaviors.ParticleEmitterCID
}

func (pc *ParticleController) Methods() ecs.ControllerMethod {
	return ecs.ControllerFrame
}

func (pc *ParticleController) Target(target ecs.Component, e ecs.Entity) bool {
	pc.Entity = e
	pc.ParticleEmitter = target.(*behaviors.ParticleEmitter)
	if !pc.ParticleEmitter.IsActive() {
		return false
	}
	pc.Body = core.GetBody(pc.Entity)
	if pc.Body == nil || !pc.Body.IsActive() {
		return false
	}
	pc.Spawner = behaviors.GetSpawner(pc.ParticleEmitter.Spawner)
	if pc.Spawner != nil && !pc.Spawner.IsActive() {
		return false
	}
	return true
}

func (pc *ParticleController) spawn() {
	// TODO: Optimize this
	e := Spawn(pc.Spawner)

	if e == 0 {
		return
	}

	flags := ecs.ComponentActive | ecs.ComponentHideEntityInEditor | ecs.ComponentLockedInEditor

	body := core.GetBody(e)
	if body == nil {
		// Each particle has its own position
		body = ecs.NewAttachedComponent(e, core.BodyCID).(*core.Body)
		body.Flags |= flags
	}
	vis := materials.GetVisible(e)
	if vis == nil {
		// Each particle has its own opacity
		vis = ecs.NewAttachedComponent(e, materials.VisibleCID).(*materials.Visible)
		vis.Flags |= flags
	}
	eph := behaviors.GetEphemeral(e)
	if eph == nil {
		// Each particle has its own lifetime
		eph = ecs.NewAttachedComponent(e, behaviors.EphemeralCID).(*behaviors.Ephemeral)
		eph.Flags |= flags
		eph.Lifetime = pc.Lifetime
		eph.FadeTime = pc.FadeTime
		eph.DeleteEntityOnExpiry = true
	}
	mobile := core.GetMobile(e)
	if mobile == nil {
		// Each particle has its own dynamics
		mobile = ecs.NewAttachedComponent(e, core.MobileCID).(*core.Mobile)
		mobile.Flags |= flags
		mobile.Mass = 0.25
		mobile.CrBody = core.CollideNone
		mobile.CrPlayer = core.CollideNone
		mobile.CrWall = core.CollideBounce
	}

	body.Pos.SetAll(pc.Body.Pos.Now)
	hAngle := pc.Body.Angle.Now + (rand.Float64()-0.5)*pc.XYSpread
	vAngle := (rand.Float64() - 0.5) * pc.ZSpread
	mobile.Vel.Spawn[0] = math.Cos(hAngle*concepts.Deg2rad) * math.Cos(vAngle*concepts.Deg2rad) * pc.Vel
	mobile.Vel.Spawn[1] = math.Sin(hAngle*concepts.Deg2rad) * math.Cos(vAngle*concepts.Deg2rad) * pc.Vel
	mobile.Vel.Spawn[2] = math.Sin(vAngle*concepts.Deg2rad) * pc.Vel
	mobile.Vel.ResetToSpawn()
	eph.CreationTime = ecs.Simulation.SimTimestamp

	if pc.BodyController.Target(body, e) {
		pc.BodyController.Enter(pc.Body.Sector())
		pc.BodyController.Precompute()
	}
}

func (pc *ParticleController) Frame() {
	if pc.Spawner == nil || len(pc.Spawner.Spawned) > pc.Limit {
		return
	}
	// Our goal is to spawn ~pc.Limit particles across pc.Lifetime.
	// Approximate the probability that we should spawn a particle this frame.
	// For example, for 10 particles over 10,000ms, we would get:
	// 10/(10000/7.8) = 0.0078
	probability := float64(pc.Limit) / (pc.Lifetime / constants.TimeStepMS)
	// if this probability is > 1, we need to spawn more than one particle per
	// frame.
	iterations := int(probability) + 1
	probability /= float64(iterations)
	for range iterations {
		if rand.Float64() > probability {
			return
		}
		pc.spawn()
	}
}
