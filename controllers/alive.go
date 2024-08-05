// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/concepts"
)

type AliveController struct {
	concepts.BaseController
	*behaviors.Alive
}

func init() {
	concepts.DbTypes().RegisterController(&AliveController{})
}

func (a *AliveController) ComponentIndex() int {
	return behaviors.AliveComponentIndex
}

func (a *AliveController) Methods() concepts.ControllerMethod {
	return concepts.ControllerAlways
}

func (a *AliveController) Target(target concepts.Attachable) bool {
	a.Alive = target.(*behaviors.Alive)
	return a.IsActive()
}

func (a *AliveController) Always() {
	for source, d := range a.Damages {
		if d.Cooldown.Now <= 0 {
			d.Cooldown.Detach(a.DB.Simulation)
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
