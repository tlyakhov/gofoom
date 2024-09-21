// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

type PvsController struct {
	ecs.BaseController
	*core.Sector
}

func init() {
	// Should run after the SectorController, which recalculates normals etc
	ecs.Types().RegisterController(&PvsController{}, 60)
}

func (pvs *PvsController) ComponentID() ecs.ComponentID {
	return core.SectorCID
}

func (pvs *PvsController) Methods() ecs.ControllerMethod {
	return ecs.ControllerRecalculate | ecs.ControllerLoaded
}

func (pvs *PvsController) Target(target ecs.Attachable) bool {
	pvs.Sector = target.(*core.Sector)
	return pvs.Sector.IsActive()
}

func (pvs *PvsController) Recalculate() {
	col := ecs.ColumnFor[core.InternalSegment](pvs.ECS, core.InternalSegmentCID)
	for i := range col.Cap() {
		if seg := col.Value(i); seg != nil {
			seg.AttachToSectors()
		}
	}
	pvs.updatePVS(make([]*concepts.Vector2, 0), nil, nil, nil)
}

func (pvs *PvsController) Loaded() {
	pvs.Recalculate()
}

// TODO: There's a bug with dynamic lights: how/when do we update the PVL?
func (pvs *PvsController) updatePVS(normals []*concepts.Vector2, visitor *core.Sector, min, max *concepts.Vector3) {
	if visitor == nil {
		pvs.Sector.PVS = make(map[ecs.Entity]*core.Sector)
		pvs.Sector.PVL = make(map[ecs.Entity]*core.Body)
		pvs.Sector.PVS[pvs.Entity] = pvs.Sector
		visitor = pvs.Sector
	}

	for entity, body := range visitor.Bodies {
		if core.GetLight(body.ECS, entity) != nil {
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

		adjsec := core.GetSector(pvs.Sector.ECS, seg.AdjacentSector)
		pvs.Sector.PVS[seg.AdjacentSector] = adjsec

		normals[nNormals] = &seg.Normal
		pvs.updatePVS(normals, adjsec, adjmin, adjmax)
	}
}
