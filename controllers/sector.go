// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

type SectorController struct {
	ecs.BaseController
	*core.Sector
	pvsRecalculated int
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

func (sc *SectorController) Target(target ecs.Attachable) bool {
	sc.Sector = target.(*core.Sector)
	return sc.Sector.IsActive()
}

func (sc *SectorController) Recalculate() {
	sc.Sector.Recalculate()
}

// TODO: This should be done with an actual queue, to avoid big lag spikes
func (sc *SectorController) pvs() {
	// Only update a few sectors
	if sc.pvsRecalculated > 4 {
		return
	}
	// Tolerate 10 frame refresh?
	if sc.ECS.Frame-sc.LastPVSRefresh < 10 {
		return
	}

	sc.pvsRecalculated++
	sc.LastPVSRefresh = sc.ECS.Frame
	updatePVS(sc.Sector, make([]*concepts.Vector2, 0), nil, nil, nil)
}

func (sc *SectorController) Always() {
	frame := sc.LastSeenFrame.Load()
	// This sector hasn't been observed recently
	// TODO: In this case, we should queue up a PVS refresh at the tail of the queue.
	if frame <= 0 {
		return
	}
	// This sector has been observed, queue up recalculating PVS
	sc.pvs()
	// Cache for a maximum number of frames
	if sc.ECS.Frame-uint64(frame) < 120 {
		return
	}
	sc.Lightmap.Clear()
	sc.LastSeenFrame.Store(-1)
}
