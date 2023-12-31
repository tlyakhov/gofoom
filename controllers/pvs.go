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
	pvs.updateBodyPVS(new(concepts.Vector2), nil, nil, nil)
}

func (pvs *PvsController) Loaded() {
	pvs.Recalculate()
}

func (pvs *PvsController) updateBodyPVS(normal *concepts.Vector2, visitor *core.Sector, min, max *concepts.Vector3) {
	if visitor == nil {
		pvs.Sector.PVS = make(map[uint64]*core.Sector)
		pvs.Sector.PVL = make(map[uint64]*concepts.EntityRef)
		pvs.Sector.PVS[pvs.Entity] = pvs.Sector
		visitor = pvs.Sector
	}

	for entity, body := range visitor.Bodies {
		if body.Component(core.LightComponentIndex) != nil {
			pvs.Sector.PVL[entity] = body
		}
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
		if !correctSide || pvs.Sector.PVS[adj.Sector.Entity] != nil {
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
		pvs.Sector.PVS[seg.AdjacentSector.Entity] = adjsec

		if normal.Zero() {
			pvs.updateBodyPVS(&seg.Normal, adjsec, adjmin, adjmax)
		} else {
			pvs.updateBodyPVS(normal, adjsec, adjmin, adjmax)
		}
	}
}
