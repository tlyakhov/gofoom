// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/ecs"
)

type AliveController struct {
	ecs.BaseController
	*behaviors.Alive
}

func init() {
	ecs.Types().RegisterController(&AliveController{}, 100)
}

func (a *AliveController) ComponentID() ecs.ComponentID {
	return behaviors.AliveCID
}

func (a *AliveController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways
}

func (a *AliveController) Target(target ecs.Attachable) bool {
	a.Alive = target.(*behaviors.Alive)
	return a.IsActive()
}

func (a *AliveController) Always() {
	for source, d := range a.Damages {
		if d.Cooldown.Now <= 0 {
			d.Cooldown.Detach(a.ECS.Simulation)
			delete(a.Damages, source)
			continue
		}
		if d.Amount > 0 {
			if a.Health > 0 {
				a.Health -= d.Amount
			}
			d.Amount = 0
		}
		d.Cooldown.Now--
	}
}
