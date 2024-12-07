// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
)

type InternalSegmentController struct {
	ecs.BaseController
	*core.InternalSegment
}

func init() {
	// Should run after the SectorController, which recalculates normals etc
	ecs.Types().RegisterController(func() ecs.Controller { return &InternalSegmentController{} }, 60)
}

func (isc *InternalSegmentController) ComponentID() ecs.ComponentID {
	return core.InternalSegmentCID
}

func (isc *InternalSegmentController) Methods() ecs.ControllerMethod {
	return ecs.ControllerRecalculate
}

func (isc *InternalSegmentController) Target(target ecs.Attachable, e ecs.Entity) bool {
	isc.Entity = e
	isc.InternalSegment = target.(*core.InternalSegment)
	return isc.InternalSegment.IsActive()
}

func (isc *InternalSegmentController) Recalculate() {
	isc.AttachToSectors()
}
