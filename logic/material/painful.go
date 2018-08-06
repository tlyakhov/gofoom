package logic

import (
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/entities"
	"github.com/tlyakhov/gofoom/logic/entity"
	"github.com/tlyakhov/gofoom/materials"
)

type PainfulService struct {
	*materials.Painful
}

func NewPainfulService(m *materials.Painful) *PainfulService {
	return &PainfulService{Painful: m}
}

func (m *PainfulService) ActOnEntity(e core.AbstractEntity) {
	if m.Hurt == 0 {
		return
	}

	if ae, ok := e.(*entities.AliveEntity); ok && ae.HurtTime == 0 {
		entity.NewAliveEntityService(ae).Hurt(m.Hurt)
	}
}
