// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
)

type SectorController struct {
	ecs.BaseController
	*core.Sector
}

func init() {
	// Should run before everything
	ecs.Types().RegisterController(&SectorController{}, 50)
}

func (sc *SectorController) ComponentIndex() int {
	return core.SectorComponentIndex
}

func (sc *SectorController) Methods() ecs.ControllerMethod {
	return ecs.ControllerRecalculate
}

func (sc *SectorController) Target(target ecs.Attachable) bool {
	sc.Sector = target.(*core.Sector)
	return sc.Sector.IsActive()
}

func (sc *SectorController) Recalculate() {
	sc.Sector.Recalculate()
	sc.Sector.LightmapBias[0] = math.MaxInt64
}
