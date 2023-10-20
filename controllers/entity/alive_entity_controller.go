package entity

import (
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/entities"
)

type AliveEntityController struct {
	*PhysicalEntityController
	*entities.AliveEntity
}

func NewAliveEntityController(ae *entities.AliveEntity, e core.AbstractEntity) *AliveEntityController {
	return &AliveEntityController{AliveEntity: ae, PhysicalEntityController: NewPhysicalEntityController(&ae.PhysicalEntity, e)}
}

func (e *AliveEntityController) Hurt(amount float64) {
	e.Health -= amount
}

func (e *AliveEntityController) HurtTime() float64 {
	return e.AliveEntity.HurtTime
}
