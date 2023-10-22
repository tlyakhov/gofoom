package sector

import (
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/sectors"
)

type VerticalDoorController struct {
	*PhysicalSectorController
	*sectors.VerticalDoor
}

func NewVerticalDoorController(s *sectors.VerticalDoor) *VerticalDoorController {
	return &VerticalDoorController{VerticalDoor: s, PhysicalSectorController: NewPhysicalSectorController(&s.PhysicalSector)}
}

func (s *VerticalDoorController) ActOnMob(e core.AbstractMob) {
	s.PhysicalSectorController.ActOnMob(e)

	ps := s.PhysicalSectorController

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

func (s *VerticalDoorController) Frame() {
	ps := s.PhysicalSectorController
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

func (s *VerticalDoorController) Recalculate() {
	s.PhysicalSectorController.PhysicalSector.Recalculate()
	s.PhysicalSectorController.Max[2] = s.PhysicalSectorController.TopZ.Original
	s.UpdatePVS()
}
