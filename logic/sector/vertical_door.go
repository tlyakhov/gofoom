package sector

import (
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/sectors"
)

type VerticalDoorService struct {
	*PhysicalSectorService
	*sectors.VerticalDoor
}

func NewVerticalDoorService(s *sectors.VerticalDoor) *VerticalDoorService {
	return &VerticalDoorService{VerticalDoor: s, PhysicalSectorService: NewPhysicalSectorService(&s.PhysicalSector)}
}

func (s *VerticalDoorService) ActOnEntity(e core.AbstractEntity) {
	s.PhysicalSectorService.ActOnEntity(e)

	//if _, ok := e.(*LightEntity); ok {
	//	return
	//}

	ps := s.PhysicalSectorService

	if ps.Center.Dist(e.Physical().Pos) < 100 {
		if s.State == sectors.Closed || s.State == sectors.Closing {
			s.State = sectors.Opening
			s.VelZ = constants.DoorSpeed
		}
	} else if s.State == sectors.Open {
		s.State = sectors.Closing
		s.VelZ = -constants.DoorSpeed
	}
}

func (s *VerticalDoorService) Frame(lastFrameTime float64) {
	ps := s.PhysicalSectorService

	last := ps.TopZ
	ps.TopZ += s.VelZ * lastFrameTime / 30.0

	if ps.TopZ < ps.BottomZ {
		ps.TopZ = ps.BottomZ
		s.VelZ = 0
		s.State = sectors.Closed
	}
	if ps.TopZ > s.OrigTopZ {
		ps.TopZ = s.OrigTopZ
		s.VelZ = 0
		s.State = sectors.Open
	}

	if last != ps.TopZ {
		ps.ClearLightmaps()
		//for (var key in this.pvs) {
		//   this.pvs[key].clearLightmaps();
		//}
	}
}
