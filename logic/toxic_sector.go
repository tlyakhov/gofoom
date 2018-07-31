package logic

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/mapping"
)

type ToxicSector mapping.ToxicSector

func (s *ToxicSector) Collide(e *Entity) {
	concepts.Local(s.Sector, TypeMap).(*Sector).Collide(e)
	ae, ok := ((concepts.ISerializable)(e)).(*AliveEntity)
	if ok && s.Hurt != 0 && ae.HurtTime == 0 {
		ae.Hurt(s.Hurt)
	}

	if s.FloorMaterial != nil && e.Pos.Z <= s.BottomZ {
		m := concepts.Local(s.FloorMaterial, TypeMap).(*Painful)
		m.ActOnEntity(e)
	}
	if s.CeilMaterial != nil && e.Pos.Z >= s.TopZ {
		m := concepts.Local(s.CeilMaterial, TypeMap).(*Painful)
		m.ActOnEntity(e)
	}
}
