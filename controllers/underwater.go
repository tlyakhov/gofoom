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
	*behaviors.Underwater
	Sector *core.Sector
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &UnderwaterController{} }, 100)
}

func (uc *UnderwaterController) ComponentID() ecs.ComponentID {
	return behaviors.UnderwaterCID
}

func (uc *UnderwaterController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways | ecs.ControllerRecalculate
}

func (uc *UnderwaterController) Target(target ecs.Attachable, e ecs.Entity) bool {
	uc.Entity = e
	uc.Underwater = target.(*behaviors.Underwater)
	uc.Sector = core.GetSector(uc.Entity)
	return uc.IsActive() && uc.Sector.IsActive()
}

func (uc *UnderwaterController) Always() {
	for e := range uc.Sector.Bodies {
		if mobile := core.GetMobile(e); mobile != nil {
			mobile.Vel.Now.MulSelf(1.0 / constants.SwimDamping)
		}
	}
}

func (uc *UnderwaterController) Recalculate() {
	// TODO: This has a code smell. Should be set by the user
	uc.Sector.Gravity = concepts.Vector3{0, 0, -constants.GravitySwim}
}
