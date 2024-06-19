// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

type SectorController struct {
	concepts.BaseController
	*core.Sector
}

func init() {
	concepts.DbTypes().RegisterController(&SectorController{})
}

func (sc *SectorController) ComponentIndex() int {
	return core.SectorComponentIndex
}

// Should run before everything
func (a *SectorController) Priority() int {
	return 50
}

func (a *SectorController) Methods() concepts.ControllerMethod {
	return concepts.ControllerRecalculate
}

func (a *SectorController) Target(target concepts.Attachable) bool {
	a.Sector = target.(*core.Sector)
	return a.Sector.IsActive()
}

func (a *SectorController) Recalculate() {
	a.Sector.Recalculate()
}
