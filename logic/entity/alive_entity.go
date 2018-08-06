package entity

import (
	"github.com/tlyakhov/gofoom/entities"
)

type AliveEntityService struct {
	*PhysicalEntityService
	*entities.AliveEntity
}

func NewAliveEntityService(e *entities.AliveEntity) *AliveEntityService {
	return &AliveEntityService{AliveEntity: e, PhysicalEntityService: NewPhysicalEntityService(&e.PhysicalEntity)}
}

func (e *AliveEntityService) Hurt(amount float64) {
	e.Health -= amount
}
