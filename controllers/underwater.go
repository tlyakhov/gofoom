package controllers

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/sectors"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type UnderwaterController struct {
	concepts.BaseController
	Underwater *sectors.Underwater
	Sector     *core.Sector
}

func init() {
	concepts.DbTypes().RegisterController(&UnderwaterController{})
}

func (uc *UnderwaterController) Methods() concepts.ControllerMethod {
	return concepts.ControllerAlways | concepts.ControllerLoaded
}

func (uc *UnderwaterController) Target(target *concepts.EntityRef) bool {
	uc.TargetEntity = target
	uc.Underwater = sectors.UnderwaterFromDb(target)
	uc.Sector = core.SectorFromDb(target)
	return uc.Underwater != nil && uc.Underwater.Active && uc.Sector != nil && uc.Sector.Active
}

func (uc *UnderwaterController) Always() {
	for _, ref := range uc.Sector.Bodies {
		body := core.BodyFromDb(ref)
		if body == nil {
			continue
		}
		body.Vel.Now.MulSelf(1.0 / constants.SwimDamping)
	}
}

func (uc *UnderwaterController) Loaded() {
	uc.Sector.Gravity = concepts.Vector3{0, 0, -constants.GravitySwim}
}
