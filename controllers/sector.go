// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

type SectorController struct {
	concepts.BaseController
	*core.Sector
}

func init() {
	// Should run before everything
	concepts.DbTypes().RegisterController(&SectorController{}, 50)
}

func (sc *SectorController) ComponentIndex() int {
	return core.SectorComponentIndex
}

func (sc *SectorController) Methods() concepts.ControllerMethod {
	return concepts.ControllerRecalculate
}

func (sc *SectorController) Target(target concepts.Attachable) bool {
	sc.Sector = target.(*core.Sector)
	return sc.Sector.IsActive()
}

func (sc *SectorController) Recalculate() {
	sc.Sector.Recalculate()
	sc.Sector.LightmapBias[0] = math.MaxInt64
}
