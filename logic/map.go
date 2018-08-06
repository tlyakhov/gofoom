package logic

import (
	"github.com/tlyakhov/gofoom/logic/provide"
	"github.com/tlyakhov/gofoom/mapping"
)

type MapService struct {
	*mapping.Map
}

func NewMapService(m *mapping.Map) *MapService {
	return &MapService{Map: m}
}

func (m *MapService) Frame(lastFrameTime float64) {
	player := provide.EntityAnimator.For(m.Player)
	player.Frame(lastFrameTime)

	for _, sector := range m.Sectors {
		provide.Interactor.For(sector).ActOnEntity(m.Player)
		for _, e := range sector.GetSector().Entities {
			if !e.GetEntity().Active {
				continue
			}
			for _, pvs := range sector.GetSector().PVSEntity {
				_ = pvs
				//pvs.ActOnEntity(e)
			}
		}
		provide.SectorAnimator.For(sector).Frame(lastFrameTime)
	}
}
