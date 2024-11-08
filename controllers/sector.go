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
	pvsQueue *core.PvsQueue
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
	if len(sc.Sector.PVS) == 0 {
		updatePVS(sc.Sector, make([]*concepts.Vector2, 0), nil, nil, nil)
	}
}

func (sc *SectorController) getOrCreateQ() {
	if sc.pvsQueue != nil {
		return
	}
	a := sc.ECS.First(core.PvsQueueCID)
	if a != nil {
		sc.pvsQueue = a.(*core.PvsQueue)
	} else {
		e := sc.ECS.NewEntity()
		sc.pvsQueue = sc.ECS.NewAttachedComponent(e, core.PvsQueueCID).(*core.PvsQueue)
		sc.pvsQueue.System = true
	}
}

func (sc *SectorController) Always() {
	sc.getOrCreateQ()
	frame := sc.LastSeenFrame.Load()
	// This sector hasn't been observed recently
	if frame <= 0 {
		sc.pvsQueue.PushTail(sc.Sector)
		return
	}
	// This sector has been observed, queue up recalculating PVS
	// Tolerate 10 frame refresh?
	if sc.ECS.Frame-sc.LastPVSRefresh > 10 {
		sc.pvsQueue.PushHead(sc.Sector)
	}

	// Cache for a maximum number of frames
	if sc.ECS.Frame-uint64(frame) < 120 {
		return
	}
	sc.Lightmap.Clear()
	sc.LastSeenFrame.Store(-1)
}
