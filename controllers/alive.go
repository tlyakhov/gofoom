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

func (a *AliveController) Methods() concepts.ControllerMethod {
	return concepts.ControllerAlways
}

func (a *AliveController) Target(target *concepts.EntityRef) bool {
	a.TargetEntity = target
	a.Alive = behaviors.AliveFromDb(target)
	return a.Alive != nil && a.Alive.Active
}

func (a *AliveController) Always() {
	for source, d := range a.Damages {
		if d.Cooldown.Now <= 0 {
			d.Cooldown.Detach(a.Simulation)
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
