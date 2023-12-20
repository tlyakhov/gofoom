package controllers

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

type PvsController struct {
	concepts.BaseController
	*core.Sector
}

func init() {
	concepts.DbTypes().RegisterController(&PvsController{})
}

// Should run after the SectorController, which recalculates normals etc
func (pvs *PvsController) Priority() int {
	return 60
}

func (pvs *PvsController) Methods() concepts.ControllerMethod {
	return concepts.ControllerRecalculate | concepts.ControllerLoaded
}

func (pvs *PvsController) Target(target *concepts.EntityRef) bool {
	pvs.TargetEntity = target
	pvs.Sector = core.SectorFromDb(target)
	return pvs.Sector != nil && pvs.Sector.Active
}

func (pvs *PvsController) Recalculate() {
	pvs.buildPVS(nil)
	pvs.updateBodyPVS(new(concepts.Vector2), nil, nil, nil)
}

func (pvs *PvsController) Loaded() {
	pvs.Recalculate()
}

func (pvs *PvsController) occludedBy(visitor *core.Sector) bool {
	return false
	// Check if the "visitor" sector is completely blocked by a non-portal- or zero-height-portal segment.
	// Performance of this is terrible... :(
	// For a map of 10000 segments & current sector = 10 segs, this loop could run:
	// 10 * 10000 * 10000 = 1B times
	// This loop is all the potential occluding sectors.

	// This loop is for our visitor segments
	for _, vseg := range visitor.Segments {
		// Then our target sector segments
		for _, oseg := range pvs.Sector.Segments {
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

			for _, isector := range pvs.DB.All(core.SectorComponentIndex) {
				if isector.Ref().Entity == pvs.Sector.Entity || isector.Ref().Entity == visitor.Entity {
					continue
				}
				sector := isector.(*core.Sector)
				for _, iseg := range sector.Segments {
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

func (pvs *PvsController) buildPVS(visitor *core.Sector) {
	if visitor == nil {
		pvs.Sector.PVS = make(map[uint64]*core.Sector)
		pvs.Sector.PVS[pvs.Sector.Entity] = pvs.Sector
		pvs.Sector.PVL = make(map[uint64]*concepts.EntityRef)
		visitor = pvs.Sector
	} else if pvs.occludedBy(visitor) {
		return
	}

	pvs.Sector.PVS[visitor.Entity] = visitor

	for entity, body := range visitor.Bodies {
		if body.Component(core.LightComponentIndex) != nil {
			pvs.Sector.PVL[entity] = body
		}
	}

	for _, seg := range visitor.Segments {
		if seg.AdjacentSector.Nil() {
			continue
		}
		if pvs.Sector.PVS[seg.AdjacentSector.Entity] != nil {
			continue
		}

		adj := core.SectorFromDb(seg.AdjacentSector)

		if adj.Min[2] >= pvs.Sector.Max[2] || adj.Max[2] <= pvs.Sector.Min[2] {
			continue
		}

		pvs.buildPVS(adj)
	}
}

func (pvs *PvsController) updateBodyPVS(normal *concepts.Vector2, visitor *core.Sector, min, max *concepts.Vector3) {
	if visitor == nil {
		pvs.Sector.PVSBody = make(map[uint64]*core.Sector)
		pvs.Sector.PVSBody[pvs.Entity] = pvs.Sector
		visitor = pvs.Sector
	}
	if min == nil || max == nil {
		min, max = &pvs.Sector.Min, &pvs.Sector.Max
	}

	for _, seg := range visitor.Segments {
		adj := seg.AdjacentSegment
		if adj == nil {
			continue
		}
		correctSide := normal.Zero() || normal.Dot(&seg.Normal) >= 0
		if !correctSide || pvs.Sector.PVSBody[adj.Sector.Entity] != nil {
			continue
		}
		if adj.Sector.Min[2] >= max[2] || adj.Sector.Max[2] <= min[2] {
			continue
		}
		adjmax := max
		adjmin := min
		if adj.Sector.Max[2] < max[2] {
			adjmax = &adj.Sector.Max
		}
		if adj.Sector.Min[2] > min[2] {
			adjmin = &adj.Sector.Min
		}

		adjsec := core.SectorFromDb(seg.AdjacentSector)
		pvs.Sector.PVSBody[seg.AdjacentSector.Entity] = adjsec

		if normal.Zero() {
			pvs.updateBodyPVS(&seg.Normal, adjsec, adjmin, adjmax)
		} else {
			pvs.updateBodyPVS(normal, adjsec, adjmin, adjmax)
		}
	}
}
