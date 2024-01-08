package controllers

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/sectors"
	"tlyakhov/gofoom/concepts"
)

type VerticalDoorController struct {
	concepts.BaseController
	*sectors.VerticalDoor
	Sector *core.Sector
}

func init() {
	concepts.DbTypes().RegisterController(&VerticalDoorController{})
}

func (vd *VerticalDoorController) ComponentIndex() int {
	return sectors.VerticalDoorComponentIndex
}

func (vd *VerticalDoorController) Methods() concepts.ControllerMethod {
	return concepts.ControllerAlways
}

func (vd *VerticalDoorController) Target(target concepts.Attachable) bool {
	vd.VerticalDoor = target.(*sectors.VerticalDoor)
	vd.Sector = core.SectorFromDb(target.Ref())
	return vd.VerticalDoor.IsActive() && vd.Sector.IsActive()
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
		vd.State = sectors.DoorStateOpen
		if vd.Intent == sectors.DoorIntentOpen {
			vd.Intent = sectors.DoorIntentClosed
		}
	}
	if a.Percent >= 1 {
		vd.State = sectors.DoorStateClosed
		if vd.Intent == sectors.DoorIntentClosed {
			vd.Intent = sectors.DoorIntentReset
		}
	}

	if vd.Intent == sectors.DoorIntentOpen && vd.State != sectors.DoorStateOpen {
		vd.State = sectors.DoorStateOpening
		a.Reverse = true
	} else if vd.Intent == sectors.DoorIntentClosed && vd.State != sectors.DoorStateClosed {
		vd.State = sectors.DoorStateClosing
		a.Reverse = false
	}
}
