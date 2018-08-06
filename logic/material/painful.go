package logic

import (
	"github.com/tlyakhov/gofoom/logic/entity"
	"github.com/tlyakhov/gofoom/mapping"
	"github.com/tlyakhov/gofoom/mapping/material"
)

type PainfulService struct {
	*material.Painful
}

func NewPainfulService(m *material.Painful) *PainfulService {
	return &PainfulService{Painful: m}
}

func (m *PainfulService) ActOnEntity(e mapping.AbstractEntity) {
	if m.Hurt == 0 {
		return
	}

	if ae, ok := e.(*mapping.AliveEntity); ok && ae.HurtTime == 0 {
		entity.NewAliveEntityService(ae).Hurt(m.Hurt)
	}
}
