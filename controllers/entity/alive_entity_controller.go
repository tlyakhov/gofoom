package entity

import (
	"tlyakhov/gofoom/entities"
)

type AliveEntityController struct {
	*entities.AliveEntity
}

func NewAliveEntityController(ae *entities.AliveEntity) *AliveEntityController {
	return &AliveEntityController{AliveEntity: ae}
}

func (e *AliveEntityController) Hurt(amount float64) {
	e.Health -= amount
}

func (e *AliveEntityController) HurtTime() float64 {
	return e.AliveEntity.HurtTime
}
