package material

import (
	"tlyakhov/gofoom/controllers/provide"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/materials"
)

type PainfulController struct {
	*materials.Painful
}

func NewPainfulController(m *materials.Painful) *PainfulController {
	return &PainfulController{Painful: m}
}

func (m *PainfulController) ActOnEntity(e core.AbstractEntity) {
	if m.Hurt == 0 {
		return
	}
	hurter, ok := provide.Hurter.For(e)
	if ok && hurter.HurtTime() == 0 {
		hurter.Hurt(m.Hurt)
	}
}
