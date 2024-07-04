// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

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

func (pvs *PvsController) ComponentIndex() int {
	return core.SectorComponentIndex
}

// Should run after the SectorController, which recalculates normals etc
func (pvs *PvsController) Priority() int {
	return 60
}

func (pvs *PvsController) Methods() concepts.ControllerMethod {
	return concepts.ControllerRecalculate | concepts.ControllerLoaded
}

func (pvs *PvsController) Target(target concepts.Attachable) bool {
	pvs.Sector = target.(*core.Sector)
	return pvs.Sector.IsActive()
}

func (pvs *PvsController) Recalculate() {
	for _, attachable := range pvs.DB.AllOfType(core.InternalSegmentComponentIndex) {
		seg := attachable.(*core.InternalSegment)
		seg.AttachToSectors()
	}
	pvs.updatePVS(make([]*concepts.Vector2, 0), nil, nil, nil)
}

func (pvs *PvsController) Loaded() {
	pvs.Recalculate()
}

func (pvs *PvsController) updatePVS(normals []*concepts.Vector2, visitor *core.Sector, min, max *concepts.Vector3) {
	if visitor == nil {
		pvs.Sector.PVS = make(map[concepts.Entity]*core.Sector)
		pvs.Sector.PVL = make(map[concepts.Entity]*core.Body)
		pvs.Sector.PVS[pvs.Entity] = pvs.Sector
		visitor = pvs.Sector
	}

	for entity, body := range visitor.Bodies {
		if core.LightFromDb(body.DB, entity) != nil {
			pvs.Sector.PVL[entity] = body
		}
	}

	if min == nil || max == nil {
		min, max = &pvs.Sector.Min, &pvs.Sector.Max
	}
	nNormals := len(normals)
	normals = append(normals, nil)

	for _, seg := range visitor.Segments {
		adj := seg.AdjacentSegment
		if adj == nil {
			continue
		}
		correctSide := true
		for _, normal := range normals[:nNormals] {
			correctSide = correctSide && normal.Dot(&seg.Normal) >= 0
		}
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

		adjsec := core.SectorFromDb(pvs.Sector.DB, seg.AdjacentSector)
		pvs.Sector.PVS[seg.AdjacentSector] = adjsec

		normals[nNormals] = &seg.Normal
		pvs.updatePVS(normals, adjsec, adjmin, adjmax)
	}
}
