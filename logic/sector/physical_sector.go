package sector

import (
	"math"

	"tlyakhov/gofoom/behaviors"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/logic/provide"
)

type PhysicalSectorService struct {
	*core.PhysicalSector
}

func NewPhysicalSectorService(s *core.PhysicalSector) *PhysicalSectorService {
	return &PhysicalSectorService{PhysicalSector: s}
}

func (s *PhysicalSectorService) OnEnter(e core.AbstractEntity) {
	p := &e.Physical().Pos.Now
	if s.FloorTarget == nil && p[2] <= e.GetSector().Physical().BottomZ {
		p[2] = e.GetSector().Physical().BottomZ
	}
}

func (s *PhysicalSectorService) OnExit(e core.AbstractEntity) {
}

func (s *PhysicalSectorService) Collide(e core.AbstractEntity) {
	concrete := e.Physical()
	entityTop := concrete.Pos.Now[2] + concrete.Height
	floorZ, ceilZ := s.CalcFloorCeilingZ(concrete.Pos.Now.To2D())

	concrete.OnGround = false
	if s.FloorTarget != nil && entityTop < floorZ {
		provide.Passer.For(concrete.Sector).OnExit(e)
		concrete.Sector = s.FloorTarget
		provide.Passer.For(concrete.Sector).OnEnter(e)
		_, ceilZ = concrete.Sector.Physical().CalcFloorCeilingZ(concrete.Pos.Now.To2D())
		concrete.Pos.Now[2] = ceilZ - concrete.Height - 1.0
	} else if s.FloorTarget != nil && concrete.Pos.Now[2] <= floorZ && concrete.Vel.Now[2] > 0 {
		concrete.Vel.Now[2] = constants.PlayerJumpStrength
	} else if s.FloorTarget == nil && concrete.Pos.Now[2] <= floorZ {
		concrete.Vel.Now[2] = 0
		concrete.Pos.Now[2] = floorZ
		concrete.OnGround = true
		//fmt.Printf("%v\n", concrete.Pos)
	}

	if s.CeilTarget != nil && entityTop > ceilZ {
		provide.Passer.For(concrete.Sector).OnExit(e)
		concrete.Sector = s.CeilTarget
		provide.Passer.For(concrete.Sector).OnEnter(e)
		floorZ, _ = concrete.Sector.Physical().CalcFloorCeilingZ(concrete.Pos.Now.To2D())
		concrete.Pos.Now[2] = floorZ - concrete.Height + 1.0
	} else if s.CeilTarget == nil && entityTop > ceilZ {
		concrete.Vel.Now[2] = 0
		concrete.Pos.Now[2] = ceilZ - concrete.Height - 1.0
	}
}

func (s *PhysicalSectorService) ActOnEntity(e core.AbstractEntity) {
	if e.GetSector() == nil || e.GetSector().GetBase().ID != s.ID {
		return
	}

	v := &e.Physical().Vel.Now
	if e.GetBase().ID == s.Map.Player.GetBase().ID {
		v[0] /= 1.2
		v[1] /= 1.2
		if math.Abs(v[0]) < 0.0001 {
			v[0] = 0
		}
		if math.Abs(v[1]) < 0.0001 {
			v[1] = 0
		}
	}

	if e.Physical().Mass > 0 {
		v[2] -= constants.Gravity / e.Physical().Mass
	}

	s.Collide(e)
}

func (s *PhysicalSectorService) Frame(sim *core.Simulation) {
	for _, e := range s.Entities {
		if e.GetBase().ID == s.Map.Player.GetBase().ID || s.Map.EntitiesPaused {
			continue
		}
		provide.EntityAnimator.For(e).Frame(sim)
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

func (s *PhysicalSectorService) occludedBy(visitor core.AbstractSector) bool {
	//return false
	// Check if the "visitor" sector is completely blocked by a non-portal- or zero-height-portal segment.
	vphys := visitor.Physical()
	// Performance of this is terrible... :(
	// For a map of 10000 segments & current sector = 10 segs, this loop could run:
	// 10 * 10000 * 10000 = 1B times
	// This loop is all the potential occluding sectors.

	// This loop is for our visitor segments
	for _, vseg := range vphys.Segments {
		// Then our target sector segments
		for _, oseg := range s.PhysicalSector.Segments {
			if oseg.Matches(vseg) {
				continue
			}
			// We make two lines on either side and see if there is a segment that intersects both of them
			// (which means vseg is fully occluded from oseg)
			l1a := &oseg.P
			l1b := &vseg.P
			l2a := &oseg.Next.P
			l2b := &vseg.Next.P
			sameFacing := oseg.Normal.Dot(&vseg.Normal) >= 0
			if !sameFacing {
				l1b, l2b = l2b, l1b
			}

			occluded := false

			for id, isector := range s.Map.Sectors {
				if id == s.ID || id == vphys.ID {
					continue
				}
				for _, iseg := range isector.Physical().Segments {
					if iseg.AdjacentSector != nil {
						continue
					}
					isect1 := iseg.Intersect2D(l1a, l1b, new(concepts.Vector2))
					if !isect1 {
						continue
					}
					isect2 := iseg.Intersect2D(l2a, l2b, new(concepts.Vector2))
					if isect2 {
						occluded = true
						break
					}
				}
				if occluded {
					break
				}
			}

			if !occluded {
				return false
			}
		}
	}
	return true
}

func (s *PhysicalSectorService) buildPVS(visitor core.AbstractSector) {
	if visitor == nil {
		s.PVS = make(map[string]core.AbstractSector)
		s.PVS[s.ID] = s.PhysicalSector
		s.PVL = make(map[string]core.AbstractEntity)
		visitor = s.PhysicalSector
	} else if s.occludedBy(visitor) {
		return
	}

	s.PVS[visitor.GetBase().ID] = visitor

	for id, e := range visitor.Physical().Entities {
		if hasLightBehavior(e) {
			s.PVL[id] = e
		}
	}

	for _, seg := range visitor.Physical().Segments {
		adj := seg.AdjacentSector
		if adj == nil {
			continue
		}
		adjID := adj.GetBase().ID
		if s.PVS[adjID] != nil {
			continue
		}

		if adj.Physical().Min[2] >= s.PhysicalSector.Max[2] || adj.Physical().Max[2] <= s.PhysicalSector.Min[2] {
			continue
		}

		s.buildPVS(adj)
	}
}

func (s *PhysicalSectorService) updateEntityPVS(normal *concepts.Vector2, visitor core.AbstractSector) {
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
		correctSide := normal.Zero() || normal.Dot(&seg.Normal) >= 0
		if !correctSide || s.PVSEntity[adj.Sector.GetBase().ID] != nil {
			continue
		}

		s.PVSEntity[seg.AdjacentSector.GetBase().ID] = seg.AdjacentSector

		if normal.Zero() {
			s.updateEntityPVS(&seg.Normal, seg.AdjacentSector)
		} else {
			s.updateEntityPVS(normal, seg.AdjacentSector)
		}
	}
}

func (s *PhysicalSectorService) UpdatePVS() {
	s.buildPVS(nil)
	s.updateEntityPVS(new(concepts.Vector2), nil)
}

func (s *PhysicalSectorService) Recalculate() {
	s.PhysicalSector.Recalculate()
	s.UpdatePVS()
}
