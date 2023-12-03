package controllers

import (
	"image/color"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/sectors"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type UnderwaterController struct {
	concepts.BaseController
	Underwater *sectors.Underwater
	Alive      *behaviors.Alive
	Body       *core.Body
}

func init() {
	concepts.DbTypes().RegisterController(&UnderwaterController{})
}

func (uc *UnderwaterController) Source(er *concepts.EntityRef) bool {
	uc.SourceEntity = er
	uc.Underwater = sectors.UnderwaterFromDb(er)
	uc.Alive = behaviors.AliveFromDb(er)
	return uc.Underwater != nil && uc.Underwater.Active &&
		uc.Alive != nil && uc.Alive.Active
}

func (uc *UnderwaterController) Target(target *concepts.EntityRef) bool {
	uc.TargetEntity = target
	uc.Body = core.BodyFromDb(target)
	return uc.Body != nil && uc.Body.Active
}

func (uc *UnderwaterController) Containment() {
	uc.Body.Vel.Now.MulSelf(1.0 / constants.SwimDamping)
	uc.Body.Vel.Now[2] -= constants.GravitySwim
}

func (uc *UnderwaterController) Always() {
	if p := behaviors.PlayerFromDb(uc.TargetEntity); p != nil {
		p.FrameTint = color.NRGBA{75, 147, 255, 90}
	}
}

func (uc *UnderwaterController) Exit() {
	if p := behaviors.PlayerFromDb(uc.TargetEntity); p != nil {
		p.FrameTint = color.NRGBA{}
	}
}
