package sector

import (
	"math"

	"github.com/tlyakhov/gofoom/behaviors"
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/logic/provide"
)

type PhysicalSectorService struct {
	*core.PhysicalSector
}

func NewPhysicalSectorService(s *core.PhysicalSector) *PhysicalSectorService {
	return &PhysicalSectorService{PhysicalSector: s}
}

func (s *PhysicalSectorService) OnEnter(e core.AbstractEntity) {
	if s.FloorTarget == nil && e.Physical().Pos.Z <= e.GetSector().Physical().BottomZ {
		e.Physical().Pos.Z = e.GetSector().Physical().BottomZ
	}
}

func (s *PhysicalSectorService) OnExit(e core.AbstractEntity) {
}

func (s *PhysicalSectorService) Collide(e core.AbstractEntity) {
	concrete := e.Physical()
	entityTop := concrete.Pos.Z + concrete.Height
	floorZ, ceilZ := s.CalcFloorCeilingZ(concrete.Pos.To2D())

	if s.FloorTarget != nil && entityTop < floorZ {
		provide.Passer.For(concrete.Sector).OnExit(e)
		concrete.Sector = s.FloorTarget
		provide.Passer.For(concrete.Sector).OnEnter(e)
		floorZ, ceilZ = concrete.Sector.Physical().CalcFloorCeilingZ(concrete.Pos.To2D())
		concrete.Pos.Z = ceilZ - concrete.Height - 1.0
	} else if s.FloorTarget != nil && concrete.Pos.Z <= floorZ && concrete.Vel.Z > 0 {
		concrete.Vel.Z = constants.PlayerJumpStrength
	} else if s.FloorTarget == nil && concrete.Pos.Z <= floorZ {
		concrete.Vel.Z = 0
		concrete.Pos.Z = floorZ
	}

	if s.CeilTarget != nil && entityTop > ceilZ {
		provide.Passer.For(concrete.Sector).OnExit(e)
		concrete.Sector = s.CeilTarget
		provide.Passer.For(concrete.Sector).OnEnter(e)
		floorZ, ceilZ = concrete.Sector.Physical().CalcFloorCeilingZ(concrete.Pos.To2D())
		concrete.Pos.Z = floorZ - concrete.Height + 1.0
	} else if s.CeilTarget == nil && entityTop > ceilZ {
		concrete.Vel.Z = 0
		concrete.Pos.Z = ceilZ - concrete.Height - 1.0
	}
}

func (s *PhysicalSectorService) ActOnEntity(e core.AbstractEntity) {
	if e.GetSector() == nil || e.GetSector().GetBase().ID != s.ID {
		return
	}

	if e.GetBase().ID == s.Map.Player.GetBase().ID {
		e.Physical().Vel.X = 0
		e.Physical().Vel.Y = 0
	}

	e.Physical().Vel.Z -= constants.Gravity * e.Physical().Weight

	s.Collide(e)
}

func (s *PhysicalSectorService) Frame(lastFrameTime float64) {
	for _, e := range s.Entities {
		if e.GetBase().ID == s.Map.Player.GetBase().ID || s.Map.EntitiesPaused {
			continue
		}
		provide.EntityAnimator.For(e).Frame(lastFrameTime)
	}
}

func hasLightBehavior(e core.AbstractEntity) bool {
	for _, b := range e.Physical().Behaviors {
		if _, ok := b.(*behaviors.Light); ok {
			return true
		}
	}
	return false
}
func (s *PhysicalSectorService) updatePVS(normal concepts.Vector2, visitor core.AbstractSector) {
	if visitor == nil {
		visitor = s.PhysicalSector
		s.PVS = make(map[string]core.AbstractSector)
		s.PVS[s.ID] = visitor
		s.PVSLights = []core.AbstractEntity{}
		for _, e := range s.Entities {
			if hasLightBehavior(e) {
				s.PVSLights = append(s.PVSLights, e)
			}
		}
	}
	for _, seg := range visitor.Physical().Segments {
		adj := seg.AdjacentSegment
		if adj == nil {
			continue
		}

		correctSide := normal.Zero() || normal.Dot(seg.Normal) >= 0
		if !correctSide || s.PVS[seg.AdjacentSector.GetBase().ID] != nil {
			continue
		}
		s.PVS[seg.AdjacentSector.GetBase().ID] = seg.AdjacentSector

		floorZ, ceilZ := adj.Sector.Physical().CalcFloorCeilingZ(seg.P)
		if math.Abs(ceilZ-floorZ) < constants.VelocityEpsilon {
			continue
		}

		for _, e := range seg.AdjacentSector.Physical().Entities {
			if hasLightBehavior(e) {
				s.PVSLights = append(s.PVSLights, e)
			}
		}
		if normal.Zero() {
			s.updatePVS(seg.Normal, seg.AdjacentSector)
		} else {
			s.updatePVS(normal, seg.AdjacentSector)
		}
	}
}

func (s *PhysicalSectorService) updateEntityPVS(normal concepts.Vector2, visitor core.AbstractSector) {
	if visitor == nil {
		s.PVSEntity = make(map[string]core.AbstractSector)
		s.PVSEntity[s.ID] = s.PhysicalSector
		visitor = s
	}

	for _, seg := range visitor.Physical().Segments {
		adj := seg.AdjacentSegment
		if adj == nil || adj.MidMaterial != nil {
			continue
		}
		correctSide := normal.Zero() || normal.Dot(seg.Normal) >= 0
		if !correctSide || s.PVSEntity[adj.Sector.GetBase().ID] != nil {
			continue
		}

		s.PVSEntity[seg.AdjacentSector.GetBase().ID] = seg.AdjacentSector

		if normal.Zero() {
			s.updateEntityPVS(seg.Normal, seg.AdjacentSector)
		} else {
			s.updateEntityPVS(normal, seg.AdjacentSector)
		}
	}
}

func (s *PhysicalSectorService) UpdatePVS() {
	s.updatePVS(concepts.Vector2{}, nil)
	s.updateEntityPVS(concepts.Vector2{}, nil)
	s.ClearLightmaps()
}

func (s *PhysicalSectorService) Recalculate() {
	s.PhysicalSector.Recalculate()
	s.UpdatePVS()
}
