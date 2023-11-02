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
	if a.HurtTime > 0 {
		a.HurtTime-- // Should account for frame time here.
	}
}
