package logic

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/mapping"
	"github.com/tlyakhov/gofoom/registry"
)

type ToxicSector mapping.ToxicSector

func init() {
	registry.Instance().RegisterMapped(ToxicSector{}, mapping.ToxicSector{})
}

func (s *ToxicSector) Collide(e *Entity) {
	registry.Translate(s.Sector).(*Sector).Collide(e)
	ae, ok := ((concepts.ISerializable)(e)).(*AliveEntity)
	if ok && s.Hurt != 0 && ae.HurtTime == 0 {
		ae.Hurt(s.Hurt)
	}

	if s.FloorMaterial != nil && e.Pos.Z <= s.BottomZ {
		m := registry.Translate(s.FloorMaterial).(*Painful)
		m.ActOnEntity(e)
	}
	if s.CeilMaterial != nil && e.Pos.Z >= s.TopZ {
		m := registry.Translate(s.CeilMaterial).(*Painful)
		m.ActOnEntity(e)
	}
}
