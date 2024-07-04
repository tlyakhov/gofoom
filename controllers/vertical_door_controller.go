// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

type VerticalDoorController struct {
	concepts.BaseController
	*behaviors.VerticalDoor
	Sector *core.Sector
}

func init() {
	concepts.DbTypes().RegisterController(&VerticalDoorController{})
}

func (vd *VerticalDoorController) ComponentIndex() int {
	return behaviors.VerticalDoorComponentIndex
}

func (vd *VerticalDoorController) Methods() concepts.ControllerMethod {
	return concepts.ControllerAlways
}

func (vd *VerticalDoorController) Target(target concepts.Attachable) bool {
	vd.VerticalDoor = target.(*behaviors.VerticalDoor)
	vd.Sector = core.SectorFromDb(vd.VerticalDoor.DB, vd.VerticalDoor.Entity)
	return vd.VerticalDoor.IsActive() && vd.Sector != nil && vd.Sector.IsActive()
}

func (vd *VerticalDoorController) setupAnimation() {
	if vd.Sector.TopZ.Animation != nil {
		return
	}
	a := vd.Sector.TopZ.NewAnimation()
	a.SetDB(vd.DB)
	a.Construct(nil)
	a.Start = vd.Sector.TopZ.Original
	a.End = vd.Sector.BottomZ.Original
	a.Coordinates = concepts.AnimationCoordinatesAbsolute
	a.Duration = 1000
	a.TweeningFunc = concepts.EaseInOut
	a.Lifetime = concepts.AnimationLifetimeHold
}

func (vd *VerticalDoorController) Always() {
	vd.setupAnimation()

	a := vd.Sector.TopZ.Animation
	if a.Percent <= 0 {
		vd.State = behaviors.DoorStateOpen
		if vd.Intent == behaviors.DoorIntentOpen {
			vd.Intent = behaviors.DoorIntentClosed
		}
	}
	if a.Percent >= 1 {
		vd.State = behaviors.DoorStateClosed
		if vd.Intent == behaviors.DoorIntentClosed {
			vd.Intent = behaviors.DoorIntentReset
		}
	}

	if vd.Intent == behaviors.DoorIntentOpen && vd.State != behaviors.DoorStateOpen {
		vd.State = behaviors.DoorStateOpening
		a.Reverse = true
	} else if vd.Intent == behaviors.DoorIntentClosed && vd.State != behaviors.DoorStateClosed {
		vd.State = behaviors.DoorStateClosing
		a.Reverse = false
	}
}
