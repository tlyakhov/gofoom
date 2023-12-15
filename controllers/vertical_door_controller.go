package controllers

import (
	"strconv"
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

func (vd *VerticalDoorController) Methods() concepts.ControllerMethod {
	return concepts.ControllerAlways
}

func (vd *VerticalDoorController) Target(target *concepts.EntityRef) bool {
	vd.TargetEntity = target
	vd.VerticalDoor = sectors.VerticalDoorFromDb(target)
	vd.Sector = core.SectorFromDb(target)
	return vd.VerticalDoor != nil && vd.VerticalDoor.Active &&
		vd.Sector != nil && vd.Sector.Active
}

func (vd *VerticalDoorController) setupAnimation() {
	if vd.Animation != nil {
		return
	}
	name := "vd_" + strconv.FormatUint(vd.TargetEntity.Entity, 10)
	vd.Animation = new(concepts.Animation[float64])
	vd.Animation.Construct(vd.Simulation)
	vd.Animation.Name = name
	vd.Animation.Target = &vd.Sector.TopZ
	vd.Animation.Start = vd.Sector.TopZ.Original
	vd.Animation.End = vd.Sector.BottomZ.Original
	vd.Animation.Duration = 1000
	vd.Animation.EasingFunc = concepts.EaseInOut
	vd.Animation.Style = concepts.AnimationStyleHold
	vd.Animate(name, vd.Animation)
}

func (vd *VerticalDoorController) Always() {
	vd.setupAnimation()

	if vd.Animation.Percent <= 0 {
		vd.State = sectors.DoorStateOpen
		if vd.Intent == sectors.DoorIntentOpen {
			vd.Intent = sectors.DoorIntentClosed
		}
	}
	if vd.Animation.Percent >= 1 {
		vd.State = sectors.DoorStateClosed
		if vd.Intent == sectors.DoorIntentClosed {
			vd.Intent = sectors.DoorIntentReset
		}
	}

	if vd.Intent == sectors.DoorIntentOpen && vd.State != sectors.DoorStateOpen {
		vd.State = sectors.DoorStateOpening
		vd.Animation.Reverse = true
	} else if vd.Intent == sectors.DoorIntentClosed && vd.State != sectors.DoorStateClosed {
		vd.State = sectors.DoorStateClosing
		vd.Animation.Reverse = false
	}

}
