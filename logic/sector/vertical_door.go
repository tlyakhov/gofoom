package sector

import (
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/sectors"
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
		s.PhysicalSectorService.ClearLightmaps()
		for _, pvs := range ps.PVS {
			pvs.Physical().ClearLightmaps()
		}
	}
}

func (s *VerticalDoorService) Recalculate() {
	s.PhysicalSectorService.PhysicalSector.Recalculate()
	s.PhysicalSectorService.Max.Z = s.OrigTopZ
	s.UpdatePVS()
}
