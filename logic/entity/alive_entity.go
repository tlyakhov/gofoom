package entity

import (
	"github.com/tlyakhov/gofoom/mapping"
)

type AliveEntityService struct {
	*PhysicalEntityService
	*mapping.AliveEntity
}

func NewAliveEntityService(e *mapping.AliveEntity) *AliveEntityService {
	return &AliveEntityService{AliveEntity: e, PhysicalEntityService: NewPhysicalEntityService(&e.PhysicalEntity)}
}

func (e *AliveEntityService) Hurt(amount float64) {
	e.Health -= amount
}
