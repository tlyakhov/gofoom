package controllers

import (
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
		p.FrameTint = concepts.Vector4{75.0 / 255.0, 147.0 / 255.0, 1, 90.0 / 255.0}
	}
}

func (uc *UnderwaterController) Exit() {
	if p := behaviors.PlayerFromDb(uc.TargetEntity); p != nil {
		p.FrameTint = concepts.Vector4{}
	}
}
