package entity

import (
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/entities"
)

type AliveEntityService struct {
	*PhysicalEntityService
	*entities.AliveEntity
}

func NewAliveEntityService(ae *entities.AliveEntity, e core.AbstractEntity) *AliveEntityService {
	return &AliveEntityService{AliveEntity: ae, PhysicalEntityService: NewPhysicalEntityService(&ae.PhysicalEntity, e)}
}

func (e *AliveEntityService) Hurt(amount float64) {
	e.Health -= amount
}

func (e *AliveEntityService) HurtTime() float64 {
	return e.AliveEntity.HurtTime
}
