package controllers

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

type SectorController struct {
	concepts.BaseController
	*core.Sector
}

func init() {
	concepts.DbTypes().RegisterController(SectorController{})
}

func (a *SectorController) Target(target *concepts.EntityRef) bool {
	a.TargetEntity = target
	a.Sector = core.SectorFromDb(target)
	return a.Sector != nil && a.Sector.Active
}

func (a *SectorController) Loaded() {
	a.Sector.Recalculate()
	a.UpdatePVS()

	for _, er := range a.Sector.Mobs {
		a.ControllerSet.Act(&er, a.TargetEntity, "Loaded")
	}
}

func (a *SectorController) occludedBy(visitor *core.Sector) bool {
	//return false
	// Check if the "visitor" sector is completely blocked by a non-portal- or zero-height-portal segment.
	vphys := visitor
	// Performance of this is terrible... :(
	// For a map of 10000 segments & current sector = 10 segs, this loop could run:
	// 10 * 10000 * 10000 = 1B times
	// This loop is all the potential occluding sectors.

	// This loop is for our visitor segments
	for _, vseg := range vphys.Segments {
		// Then our target sector segments
		for _, oseg := range a.Sector.Segments {
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

			for entity, component := range a.DB.All(core.SectorComponentIndex) {
				if entity == a.Sector.Entity || entity == vphys.Entity {
					continue
				}
				isector := component.(*core.Sector)
				for _, iseg := range isector.Segments {
					if !iseg.AdjacentSector.Nil() {
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

func (a *SectorController) buildPVS(visitor *core.Sector) {
	if visitor == nil {
		a.Sector.PVS = make(map[uint64]*core.Sector)
		a.Sector.PVS[a.Sector.Entity] = a.Sector
		a.Sector.PVL = make(map[uint64]*concepts.EntityRef)
		visitor = a.Sector
	} else if a.occludedBy(visitor) {
		return
	}

	a.Sector.PVS[visitor.Entity] = visitor

	for entity, mob := range visitor.Mobs {
		if mob.Component(core.LightComponentIndex) != nil {
			a.Sector.PVL[entity] = &mob
		}
	}

	for _, seg := range visitor.Segments {
		if seg.AdjacentSector.Nil() {
			continue
		}
		if a.Sector.PVS[seg.AdjacentSector.Entity] != nil {
			continue
		}

		adj := core.SectorFromDb(seg.AdjacentSector)

		if adj.Min[2] >= a.Sector.Max[2] || adj.Max[2] <= a.Sector.Min[2] {
			continue
		}

		a.buildPVS(adj)
	}
}

func (a *SectorController) updateMobPVS(normal *concepts.Vector2, visitor *core.Sector) {
	if visitor == nil {
		a.Sector.PVSMob = make(map[uint64]*core.Sector)
		a.Sector.PVSMob[a.Entity] = a.Sector
		visitor = a.Sector
	}

	for _, seg := range visitor.Segments {
		adj := seg.AdjacentSegment
		if adj == nil {
			continue
		}
		correctSide := normal.Zero() || normal.Dot(&seg.Normal) >= 0
		if !correctSide || a.Sector.PVSMob[adj.Sector.Entity] != nil {
			continue
		}

		adjsec := core.SectorFromDb(seg.AdjacentSector)
		a.Sector.PVSMob[seg.AdjacentSector.Entity] = adjsec

		if normal.Zero() {
			a.updateMobPVS(&seg.Normal, adjsec)
		} else {
			a.updateMobPVS(normal, adjsec)
		}
	}
}

func (a *SectorController) UpdatePVS() {
	a.buildPVS(nil)
	a.updateMobPVS(new(concepts.Vector2), nil)
}
