package controllers

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/sectors"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
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

func (vd *VerticalDoorController) Always() {
	if vd.Intent == "open" && (vd.State == sectors.Closed || vd.State == sectors.Closing) {
		vd.State = sectors.Opening
		vd.VelZ = constants.DoorSpeed
		vd.Intent = "closed"
	}
	z := vd.Sector.TopZ.Now + vd.VelZ*constants.TimeStep
	if z < vd.Sector.BottomZ.Now {
		z = vd.Sector.BottomZ.Now
		vd.VelZ = 0
		vd.State = sectors.Closed
	}
	if z > vd.Sector.TopZ.Original {
		z = vd.Sector.TopZ.Original
		vd.VelZ = 0
		vd.State = sectors.Open
	}
	vd.Sector.TopZ.Now = z
	if vd.State == sectors.Open {
		vd.State = sectors.Closing
		vd.VelZ = -constants.DoorSpeed
	}
}
