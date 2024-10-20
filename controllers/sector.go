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

func (sc *SectorController) pvs() {
	// Only update a few sectors
	if sc.pvsRecalculated > 2 {
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
	// Cache for a maximum number of frames
	if frame <= 0 || sc.ECS.Frame-uint64(frame) < 120 {
		sc.pvs()
		return
	}
	sc.Lightmap.Clear()
	sc.LastSeenFrame.Store(-1)
}
