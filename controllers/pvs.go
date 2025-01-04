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
	*core.PvsQueue
	target *core.Sector
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &PvsController{} }, 100)
}

func (pvs *PvsController) ComponentID() ecs.ComponentID {
	return core.PvsQueueCID
}

func (pvs *PvsController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways | ecs.ControllerRecalculate
}

func (pvs *PvsController) Target(target ecs.Attachable, e ecs.Entity) bool {
	pvs.Entity = e
	pvs.PvsQueue = target.(*core.PvsQueue)
	return pvs.PvsQueue.IsActive()
}

func (pvs *PvsController) Always() {
	// We should potentially limit this further, if we have areas with a ton of sectors
	for range 2 {
		pvs.target = pvs.PopHead()
		if pvs.target == nil {
			break
		}
		pvs.updatePVS(nil, nil, nil, nil)
		// log.Printf("Refreshed pvs for %v", sector.Entity)
	}
}

func (pvs *PvsController) updatePVS(visitor *core.Sector, min, max *concepts.Vector3, normals []*concepts.Vector2) {
	// TODO: This can be very expensive for large areas with lots of connected
	// sectors. How can we optimize this?
	// TODO: Can we special-case doors to block invisible sectors when closed?
	// TODO: Should we do some kind of cheap space partitioning? Maybe a K-D Tree?
	// TODO: Simple optimization: create a PVS distance limit - if we've gone
	// through say, 5+ portals, and a sector is more than X units away from the
	// target sector, ignore it (and its portals)
	// TODO: Should we convert the map to an EntityTable at the end?
	// EntityTables have worse write performance but much better read
	// performance

	if visitor == nil {
		pvs.target.PVS = nil
		pvs.target.PVL = make([]*core.Body, 0)
		pvs.target.Colliders = make(map[ecs.Entity]*core.Mobile)
		pvs.target.PVS.Set(uint32(pvs.target.Entity))
		pvs.target.LastPVSRefresh = pvs.target.ECS.Frame
		visitor = pvs.target
	}

	for entity, body := range visitor.Bodies {
		if core.GetLight(body.ECS, entity) != nil {
			pvs.target.PVL = append(pvs.target.PVL, body)
		}
		if m := core.GetMobile(body.ECS, entity); m != nil &&
			(m.CrBody != core.CollideNone || m.CrPlayer != core.CollideNone) {
			pvs.target.Colliders[entity] = m
		}
	}

	if min == nil || max == nil {
		min, max = &pvs.target.Min, &pvs.target.Max
	}
	nNormals := len(normals)
	normals = append(normals, nil)

	for _, seg := range visitor.Segments {
		adj := seg.AdjacentSegment
		if adj == nil || seg.AdjacentSector == 0 {
			continue
		}
		if adj.Sector.Min[2] >= max[2] || adj.Sector.Max[2] <= min[2] {
			continue
		}
		if pvs.target.PVS.Contains(uint32(adj.Sector.Entity)) {
			continue
		}

		correctSide := true
		for _, normal := range normals[:nNormals] {
			correctSide = correctSide && normal.Dot(&seg.Normal) >= 0
		}
		if !correctSide {
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

		pvs.target.PVS.Set(uint32(seg.AdjacentSector))

		normals[nNormals] = &seg.Normal
		pvs.updatePVS(adj.Sector, adjmin, adjmax, normals)
	}
}
