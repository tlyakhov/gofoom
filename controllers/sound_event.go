// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"tlyakhov/gofoom/components/audio"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
)

type SoundEventController struct {
	ecs.BaseController
	*audio.SoundEvent
	Body   *core.Body
	Sector *core.Sector
	Mobile *core.Mobile
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &SoundEventController{} }, 100)
}

func (sc *SoundEventController) ComponentID() ecs.ComponentID {
	return audio.SoundEventCID
}

func (sc *SoundEventController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways
}

func (sc *SoundEventController) Target(target ecs.Component, e ecs.Entity) bool {
	sc.Entity = e
	sc.SoundEvent = target.(*audio.SoundEvent)
	if !sc.SoundEvent.IsActive() {
		return false
	}
	sc.Body = core.GetBody(sc.SourceEntity)
	sc.Sector = core.GetSector(sc.SourceEntity)
	sc.Mobile = core.GetMobile(sc.SourceEntity)
	return true
}

func (sc *SoundEventController) Always() {
	if sc.Body == nil && sc.Sector == nil {
		return
	}
	if sc.Body != nil {
		sc.SetPosition(&sc.Body.Pos.Now)
		dy, dx := math.Sincos(sc.Body.Angle.Now)
		sc.SetOrientation(&concepts.Vector3{dx * constants.UnitsPerMeter, dy * constants.UnitsPerMeter, 0})
	} else if sc.Sector != nil {
		sc.SetPosition(&sc.Sector.Center.Now)
	}

	if sc.Mobile != nil {
		sc.SetVelocity(&sc.Mobile.Vel.Now)
	}
}
