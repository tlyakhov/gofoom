// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

type pvsState struct {
	portalA *concepts.Vector2
	portalB *concepts.Vector2
}
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
		pvs.updatePVS(nil, nil, nil)
		// log.Printf("Refreshed pvs for %v", sector.Entity)
	}
}

func pvsEnsureOrder(seg *core.SectorSegment) (a *concepts.Vector2, b *concepts.Vector2) {
	// Ensure the vectors are increasing in the major axis
	a = seg.A
	b = seg.B
	if math.Abs(a[0]-b[0]) > math.Abs(a[1]-b[1]) {
		if b[0] < a[0] {
			a, b = b, a
		}
	} else {
		if b[1] < a[1] {
			a, b = b, a
		}
	}
	return
}

func (pvs *PvsController) updatePVS(visitor *core.Sector, min, max *concepts.Vector3) {
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

	/*

		1. Start at a portal
		2. loop through next portals. If
	*/

	if visitor == nil {
		pvs.target.PVS = make(map[ecs.Entity]*core.Sector)
		pvs.target.PVL = make([]*core.Body, 0)
		pvs.target.Colliders = make(map[ecs.Entity]*core.Mobile)
		pvs.target.PVS[pvs.target.Entity] = pvs.target
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

	col := ecs.ColumnFor[core.Sector](pvs.target.ECS, core.SectorCID)

	var isect concepts.Vector2
	for _, seg := range visitor.Segments {
		adj := seg.AdjacentSegment
		if adj == nil || seg.AdjacentSector == 0 {
			continue
		}
		if adj.Sector.Min[2] >= max[2] || adj.Sector.Max[2] <= min[2] {
			//	continue
		}
		/*if _, ok := pvs.visited[seg.AdjacentSector]; ok {
			continue
		}*/
		if pvs.target.PVS[adj.Sector.Entity] != nil {
			continue
		}

		if visitor != pvs.target {
			blocked := true
			for _, targetPortal := range pvs.target.Segments {
				if targetPortal.AdjacentSegment == nil && targetPortal.AdjacentSector == 0 {
					continue
				}
				blocked = false
				for i := range col.Cap() {
					sector := col.Value(i)
					if sector == nil {
						continue
					}
					for _, wall := range sector.Segments {
						if wall.AdjacentSegment != nil || wall.AdjacentSector != 0 {
							continue
						}

						if concepts.IntersectSegments(targetPortal.A, adj.A, wall.A, wall.B, &isect) != -1 &&
							concepts.IntersectSegments(targetPortal.B, adj.B, wall.A, wall.B, &isect) != -1 &&
							concepts.IntersectSegments(targetPortal.A, adj.B, wall.A, wall.B, &isect) != -1 &&
							concepts.IntersectSegments(targetPortal.B, adj.A, wall.A, wall.B, &isect) != -1 {
							blocked = true
							break
						}
					}
					if blocked {
						break
					}
				}

				if !blocked {
					break
				}
			}
			if blocked {
				continue
			}
		}

		adjmax := max
		adjmin := min
		if adj.Sector.Max[2] < max[2] {
			adjmax = &adj.Sector.Max
		}
		if adj.Sector.Min[2] > min[2] {
			adjmin = &adj.Sector.Min
		}

		pvs.target.PVS[seg.AdjacentSector] = adj.Sector

		pvs.updatePVS(adj.Sector, adjmin, adjmax)
	}
}
