package material

import (
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/logic/provide"
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
	hurter, ok := provide.Hurter.For(e)
	if ok && hurter.HurtTime() == 0 {
		hurter.Hurt(m.Hurt)
	}
}
