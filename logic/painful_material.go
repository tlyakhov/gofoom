package logic

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/mapping/material"
)

type Painful material.Painful

func (m *Painful) ActOnEntity(e concepts.ISerializable) {
	if m.Hurt == 0 {
		return
	}

	if ae, ok := e.(*AliveEntity); ok && ae.HurtTime == 0 {
		ae.Hurt(m.Hurt)
	}
}
