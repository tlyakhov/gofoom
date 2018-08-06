package entity

import (
	"github.com/tlyakhov/gofoom/mapping"
)

type AliveEntityService struct {
	*EntityService
	*mapping.AliveEntity
}

func NewAliveEntityService(e *mapping.AliveEntity) *AliveEntityService {
	return &AliveEntityService{AliveEntity: e, EntityService: NewEntityService(&e.Entity)}
}

func (e *AliveEntityService) Hurt(amount float64) {
	e.Health -= amount
}
