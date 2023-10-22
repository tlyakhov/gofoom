package mob_controllers

import (
	"tlyakhov/gofoom/mobs"
)

type AliveMobController struct {
	*mobs.AliveMob
}

func NewAliveMobController(ae *mobs.AliveMob) *AliveMobController {
	return &AliveMobController{AliveMob: ae}
}

func (e *AliveMobController) Hurt(amount float64) {
	e.Health -= amount
}

func (e *AliveMobController) HurtTime() float64 {
	return e.AliveMob.HurtTime
}
