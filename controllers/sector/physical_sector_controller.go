package sector

import (
	"log"
	"math"
	"tlyakhov/gofoom/behaviors"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/controllers/provide"
	"tlyakhov/gofoom/core"
)

type PhysicalSectorController struct {
	*core.PhysicalSector
}

func NewPhysicalSectorController(s *core.PhysicalSector) *PhysicalSectorController {
	return &PhysicalSectorController{PhysicalSector: s}
}

func (s *PhysicalSectorController) OnEnter(e core.AbstractEntity) {
	phys := e.Physical()
	phys.Sector = s.PhysicalSector.Model.(core.AbstractSector)
	s.PhysicalSector.Entities[phys.ID] = e.GetModel().(core.AbstractEntity)

	p := &phys.Pos.Now
	if s.FloorTarget == nil && p[2] <= e.GetSector().Physical().BottomZ.Now {
		p[2] = e.GetSector().Physical().BottomZ.Now
	}
}

func (s *PhysicalSectorController) OnExit(e core.AbstractEntity) {
	phys := e.Physical()
	if phys.Sector.Physical() != s.PhysicalSector {
		log.Printf("OnExit called for sector %v, but entity had %v as owner", s.PhysicalSector.ID, phys.Sector.Physical().ID)
		delete(phys.Sector.Physical().Entities, phys.ID)
	}

	delete(s.Entities, phys.ID)
}

func (s *PhysicalSectorController) Collide(e core.AbstractEntity) {
	entity := e.Physical()
	entityTop := entity.Pos.Now[2] + entity.Height
	floorZ, ceilZ := s.SlopedZNow(entity.Pos.Now.To2D())

	entity.OnGround = false
	if s.FloorTarget != nil && entityTop < floorZ {
		provide.Passer.For(entity.Sector).OnExit(e)
		provide.Passer.For(s.FloorTarget).OnEnter(e)
		_, ceilZ = entity.Sector.Physical().SlopedZNow(entity.Pos.Now.To2D())
		entity.Pos.Now[2] = ceilZ - entity.Height - 1.0
	} else if s.FloorTarget != nil && entity.Pos.Now[2] <= floorZ && entity.Vel.Now[2] > 0 {
		entity.Vel.Now[2] = constants.PlayerJumpForce
	} else if s.FloorTarget == nil && entity.Pos.Now[2] <= floorZ {
		dist := s.FloorNormal[2] * (floorZ - entity.Pos.Now[2])
		delta := s.FloorNormal.Mul(dist)
		entity.Vel.Now.AddSelf(delta)
		entity.Pos.Now.AddSelf(delta)
		entity.OnGround = true
	}

	if s.CeilTarget != nil && entityTop > ceilZ {
		provide.Passer.For(entity.Sector).OnExit(e)
		provide.Passer.For(s.CeilTarget).OnEnter(e)
		floorZ, _ = entity.Sector.Physical().SlopedZNow(entity.Pos.Now.To2D())
		entity.Pos.Now[2] = floorZ - entity.Height + 1.0
	} else if s.CeilTarget == nil && entityTop >= ceilZ {
		dist := -s.CeilNormal[2] * (entityTop - ceilZ + 1.0)
		delta := s.CeilNormal.Mul(dist)
		entity.Vel.Now.AddSelf(delta)
		entity.Pos.Now.AddSelf(delta)
	}
}

func (s *PhysicalSectorController) ActOnEntity(e core.AbstractEntity) {
	if e.GetSector() == nil || e.GetSector().GetBase().ID != s.ID {
		return
	}

	f := &e.Physical().Force
	if e.Physical().Mass > 0 {
		// Weight = g*m
		f[2] -= constants.Gravity * e.Physical().Mass
		v := &e.Physical().Vel.Now
		if !v.Zero() {
			// Air drag
			r := e.Physical().BoundingRadius * constants.MetersPerUnit
			crossSectionArea := math.Pi * r * r
			drag := concepts.Vector3{v[0], v[1], v[2]}
			drag.MulSelf(drag.Length())
			drag.MulSelf(-0.5 * constants.AirDensity * crossSectionArea * constants.SphereDragCoefficient)
			f.AddSelf(&drag)
			if e.Physical().OnGround {
				// Kinetic friction
				drag.From(v)
				g := concepts.Vector3{0, 0, constants.Gravity * e.Physical().Mass}
				drag.MulSelf(-s.FloorFriction * s.FloorNormal.Dot(&g))
				f.AddSelf(&drag)
			}
			//log.Printf("%v\n", drag)
		}
	}

	s.Collide(e)
}

func (s *PhysicalSectorController) Frame() {
	for _, e := range s.Entities {
		if e.GetBase().ID == s.Map.Player.GetBase().ID || s.Map.EntitiesPaused {
			continue
		}
		provide.EntityAnimator.For(e).Frame()
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

func (s *PhysicalSectorController) occludedBy(visitor core.AbstractSector) bool {
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

func (s *PhysicalSectorController) buildPVS(visitor core.AbstractSector) {
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

func (s *PhysicalSectorController) updateEntityPVS(normal *concepts.Vector2, visitor core.AbstractSector) {
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

func (s *PhysicalSectorController) UpdatePVS() {
	s.buildPVS(nil)
	s.updateEntityPVS(new(concepts.Vector2), nil)
}

func (s *PhysicalSectorController) Recalculate() {
	s.PhysicalSector.Recalculate()
	s.UpdatePVS()
}
