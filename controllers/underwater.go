// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
)

type UnderwaterController struct {
	ecs.BaseController
	Underwater *behaviors.Underwater
	Sector     *core.Sector
}

func init() {
	ecs.Types().RegisterController(&UnderwaterController{}, 100)
}

func (uc *UnderwaterController) ComponentIndex() int {
	return behaviors.UnderwaterComponentIndex
}

func (uc *UnderwaterController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways | ecs.ControllerLoaded
}

func (uc *UnderwaterController) Target(target ecs.Attachable) bool {
	uc.Underwater = target.(*behaviors.Underwater)
	uc.Sector = core.SectorFromDb(target.GetDB(), target.GetEntity())
	return uc.Underwater.IsActive() && uc.Sector.IsActive()
}

func (uc *UnderwaterController) Always() {
	for _, body := range uc.Sector.Bodies {
		body.Vel.Now.MulSelf(1.0 / constants.SwimDamping)
	}
}

func (uc *UnderwaterController) Loaded() {
	uc.Sector.Gravity = concepts.Vector3{0, 0, -constants.GravitySwim}
}
