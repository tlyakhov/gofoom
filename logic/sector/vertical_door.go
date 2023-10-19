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

	if ps.Center.Dist(&e.Physical().Pos.Now) < 100 {
		if s.State == sectors.Closed || s.State == sectors.Closing {
			s.State = sectors.Opening
			s.VelZ = constants.DoorSpeed
		}
	} else if s.State == sectors.Open {
		s.State = sectors.Closing
		s.VelZ = -constants.DoorSpeed
	}
}

func (s *VerticalDoorService) Frame() {
	ps := s.PhysicalSectorService
	z := ps.TopZ.Now + s.VelZ*constants.TimeStep

	if z < ps.BottomZ.Now {
		z = ps.BottomZ.Now
		s.VelZ = 0
		s.State = sectors.Closed
	}
	if z > ps.TopZ.Original {
		z = ps.TopZ.Original
		s.VelZ = 0
		s.State = sectors.Open
	}
	ps.TopZ.Now = z
}

func (s *VerticalDoorService) Recalculate() {
	s.PhysicalSectorService.PhysicalSector.Recalculate()
	s.PhysicalSectorService.Max[2] = s.PhysicalSectorService.TopZ.Original
	s.UpdatePVS()
}
