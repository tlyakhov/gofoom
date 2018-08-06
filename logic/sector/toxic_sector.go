package sector

import (
	"github.com/tlyakhov/gofoom/logic/provide"
	"github.com/tlyakhov/gofoom/mapping"
)

type ToxicSectorService struct {
	*SectorService
	*mapping.ToxicSector
}

func NewToxicSectorService(s *mapping.ToxicSector) *ToxicSectorService {
	return &ToxicSectorService{ToxicSector: s, SectorService: NewSectorService(&s.Sector)}
}

func (s *ToxicSectorService) Collide(e mapping.AbstractEntity) {
	provide.Passer.For(s.ToxicSector.Sector).Collide(e)
	ae, ok := e.(*mapping.AliveEntity)
	if ok && s.Hurt != 0 && ae.HurtTime == 0 {
		//NewAliveEntityService(ae).Hurt(s.Hurt)
		// TODO: Fix!
	}

	concrete := s.ToxicSector

	if concrete.FloorMaterial != nil && e.GetEntity().Pos.Z <= concrete.BottomZ {

		//m := registry.Translate(s.FloorMaterial, "logic").(*Painful)
		//m.ActOnEntity(e)
	}
	if concrete.CeilMaterial != nil && e.GetEntity().Pos.Z >= concrete.TopZ {
		//m := registry.Translate(s.CeilMaterial, "logic").(*Painful)
		//m.ActOnEntity(e)
	}
}
