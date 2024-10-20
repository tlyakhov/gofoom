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

func (pc *ParticleController) Target(target ecs.Attachable) bool {
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

		if shader := materials.GetShader(pc.ECS, e); shader != nil {
			shader.Stages[0].Opacity = 1.0 - concepts.Clamp((age-(pc.Lifetime-pc.FadeTime))/pc.FadeTime, 0.0, 1.0)
		}

		if age < pc.Lifetime {
			continue
		}
		toRemove = append(toRemove, e)
		pc.ECS.Delete(e)
		pc.Particles.Delete(e)
	}

	for _, e := range toRemove {
		delete(pc.Spawned, e)
	}

	if pc.Source == 0 || len(pc.Particles) > pc.Limit {
		return
	}
	if rand.Intn(10) != 0 {
		return
	}

	e := pc.ECS.NewEntity()
	pc.Particles.Add(e)
	pc.Spawned[e] = pc.ECS.Timestamp
	body := pc.ECS.NewAttachedComponent(e, core.BodyCID).(*core.Body)
	body.System = true
	mobile := pc.ECS.NewAttachedComponent(e, core.MobileCID).(*core.Mobile)
	mobile.System = true
	body.Pos.Now.From(&pc.Body.Pos.Now)
	rang := (rand.Float64() - 0.5) * 10
	mobile.Vel.Now[0] = math.Cos((pc.Body.Angle.Now+rang)*concepts.Deg2rad) * 15
	mobile.Vel.Now[1] = math.Sin((pc.Body.Angle.Now+rang)*concepts.Deg2rad) * 15
	mobile.Vel.Now[2] = (rand.Float64() - 0.5) * 10
	mobile.Mass = 0.25
	//lit := pc.ECS.NewAttachedComponent(e, materials.LitCID).(*materials.Lit)
	//lit.System = true
	/*lit.Diffuse[0] = 1.0
	lit.Diffuse[1] = 0.0
	lit.Diffuse[2] = 0.0
	lit.Diffuse[3] = 1.0*/
	linked := pc.ECS.NewAttachedComponent(e, ecs.LinkedCID).(*ecs.Linked)
	linked.System = true
	linked.Sources.Add(pc.Source)
	linked.Recalculate()
}
