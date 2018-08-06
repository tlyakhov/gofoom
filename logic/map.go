package logic

import (
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/logic/provide"
)

type MapService struct {
	*core.Map
}

func NewMapService(m *core.Map) *MapService {
	return &MapService{Map: m}
}

func (m *MapService) Frame(lastFrameTime float64) {
	player := provide.EntityAnimator.For(m.Player)
	player.Frame(lastFrameTime)

	for _, sector := range m.Sectors {
		provide.Interactor.For(sector).ActOnEntity(m.Player)
		for _, e := range sector.Physical().Entities {
			if !e.Physical().Active {
				continue
			}
			for _, pvs := range sector.Physical().PVSEntity {
				_ = pvs
				//pvs.ActOnEntity(e)
			}
		}
		provide.SectorAnimator.For(sector).Frame(lastFrameTime)
	}
}
