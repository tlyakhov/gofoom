// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

type EphemeralController struct {
	ecs.BaseController
	*behaviors.Ephemeral
	Visible *materials.Visible
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &EphemeralController{} }, 100)
}

func (pc *EphemeralController) ComponentID() ecs.ComponentID {
	return behaviors.EphemeralCID
}

func (pc *EphemeralController) Methods() ecs.ControllerMethod {
	return ecs.ControllerFrame
}

func (pc *EphemeralController) Target(target ecs.Component, e ecs.Entity) bool {
	pc.Entity = e
	pc.Ephemeral = target.(*behaviors.Ephemeral)
	if !pc.Ephemeral.IsActive() {
		return false
	}
	pc.Visible = materials.GetVisible(e)
	return true
}

func (pc *EphemeralController) Frame() {
	age := float64(ecs.Simulation.SimTimestamp-pc.CreationTime) / 1_000_000.0

	if pc.Visible != nil {
		fade := (age - (pc.Lifetime - pc.FadeTime)) / pc.FadeTime
		pc.Visible.Opacity = 1.0 - concepts.Clamp(fade, 0.0, 1.0)
	}

	if age > pc.Lifetime && pc.DeleteEntityOnExpiry {
		ecs.Delete(pc.Entity)
	}
}
