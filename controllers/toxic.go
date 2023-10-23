package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type ToxicController struct {
	concepts.BaseController
	Toxic *behaviors.Toxic
	Alive *behaviors.Alive
}

func init() {
	concepts.DbTypes().RegisterController(ToxicController{})
}

func (t *ToxicController) Source(er *concepts.EntityRef) bool {
	t.SourceEntity = er
	t.Toxic = behaviors.ToxicFromDb(er)
	return t.Toxic != nil && t.Toxic.Active
}

func (t *ToxicController) Target(target *concepts.EntityRef) bool {
	t.TargetEntity = target
	t.Alive = behaviors.AliveFromDb(target)
	return t.Alive != nil && t.Alive.Active
}

func (t *ToxicController) Contact() {
	if t.Toxic.Hurt == 0 || t.Alive.HurtTime == 0 {
		return
	}
	if t.Alive.HurtTime == 0 {
		t.Alive.Health -= t.Toxic.Hurt
		t.Alive.HurtTime = constants.PlayerHurtTime
	}
}

func (t *ToxicController) Containment() {
	t.Contact()
}
