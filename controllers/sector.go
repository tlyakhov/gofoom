// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/core"
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

func (sc *SectorController) Recalculate() {
	sc.Sector.Recalculate()
}

func (sc *SectorController) Always() {
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
