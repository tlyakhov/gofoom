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

// Should run before everything
func (a *SectorController) Priority() int {
	return 50
}

func (a *SectorController) Methods() concepts.ControllerMethod {
	return concepts.ControllerRecalculate
}

func (a *SectorController) Target(target *concepts.EntityRef) bool {
	a.TargetEntity = target
	a.Sector = core.SectorFromDb(target)
	return a.Sector != nil && a.Sector.Active
}

func (a *SectorController) Recalculate() {
	a.Sector.Recalculate()
}
