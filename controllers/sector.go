// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"log"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type SectorController struct {
	ecs.BaseController
	*core.Sector
}

func init() {
	// Should run before everything
	ecs.Types().RegisterController(func() ecs.Controller { return &SectorController{} }, 50)
}

func (sc *SectorController) ComponentID() ecs.ComponentID {
	return core.SectorCID
}

func (sc *SectorController) Methods() ecs.ControllerMethod {
	return ecs.ControllerRecalculate | ecs.ControllerAlways
}

func (sc *SectorController) Target(target ecs.Component, e ecs.Entity) bool {
	sc.Entity = e
	sc.Sector = target.(*core.Sector)
	return sc.Sector.IsActive()
}

func applySectorTransform(sector *core.Sector, d dynamic.Dynamic) {
	transform := d.(*dynamic.DynamicValue[concepts.Matrix2])
	// Transform segments if we need to
	if transform.Procedural {
		log.Printf("Input: %v, now: %v, prev: %v", transform.Input.StringHuman(), transform.Now.StringHuman(), transform.Prev.StringHuman())
	}
	if transform.Now == transform.Prev {
		return
	}
	for _, seg := range sector.Segments {
		seg.P.Now = seg.P.Spawn
		seg.P.Now[0] -= sector.Center.Spawn[0]
		seg.P.Now[1] -= sector.Center.Spawn[1]
		sector.Transform.Now.ProjectSelf(&seg.P.Now)
		seg.P.Now[0] += sector.Center.Spawn[0]
		seg.P.Now[1] += sector.Center.Spawn[1]
		seg.P.Render = seg.P.Now
	}
	sector.RecalculateNonTopological()
}

func (sc *SectorController) Recalculate() {
	sector := sc.Sector
	sc.Sector.Transform.OnPostUpdate = func(d dynamic.Dynamic) {
		applySectorTransform(sector, d)
	}
	sc.Sector.Recalculate()
}

func (sc *SectorController) TidyOverlaps(table *ecs.EntityTable) {
	if len(*table) == 0 {
		return
	}
	updated := make(ecs.EntityTable, len(*table))
	copy(updated, *table)
	for _, e := range updated {
		if e == 0 {
			continue
		}
		overlap := core.GetSector(e)
		if overlap == nil || sc.AABBIntersect(&overlap.Min, &overlap.Max, true) {
			continue
		}
		table.Delete(e)
	}
}

func (sc *SectorController) Always() {
	// Tidy every 5 frames
	if ecs.Simulation.Frame%5 == 0 {
		sc.TidyOverlaps(&sc.HigherLayers)
		sc.TidyOverlaps(&sc.LowerLayers)
	}
	// This section is for some rendering cache invalidation

	frame := sc.LastSeenFrame.Load()

	// This sector hasn't been observed recently
	if frame <= 0 {
		return
	}

	// Cache for a maximum number of frames
	if ecs.Simulation.Frame-uint64(frame) < 120 {
		return
	}
	sc.Lightmap.Clear()
	sc.LastSeenFrame.Store(-1)
}
