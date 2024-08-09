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
	return concepts.ControllerAlways | concepts.ControllerRecalculate
}

func (vd *VerticalDoorController) Target(target concepts.Attachable) bool {
	vd.VerticalDoor = target.(*behaviors.VerticalDoor)
	vd.Sector = core.SectorFromDb(vd.VerticalDoor.DB, vd.VerticalDoor.Entity)
	return vd.VerticalDoor.IsActive() && vd.Sector != nil && vd.Sector.IsActive()
}

func (vd *VerticalDoorController) setupAnimation() {
	a := vd.Sector.TopZ.NewAnimation()
	a.SetDB(vd.DB)
	a.Construct(nil)
	a.Start = vd.Sector.TopZ.Original
	a.End = vd.Sector.BottomZ.Original
	a.Coordinates = concepts.AnimationCoordinatesAbsolute
	a.Duration = vd.Duration
	a.TweeningFunc = vd.TweeningFunc
	a.Lifetime = concepts.AnimationLifetimeOnce
}

func (vd *VerticalDoorController) adjustTransforms() {
	a := vd.Sector.TopZ.Animation

	if a.Now == a.Prev {
		return
	}

	t := concepts.Matrix2{}
	t.SetIdentity()
	var v float64
	for _, seg := range vd.Sector.Segments {
		if seg.AdjacentSegment == nil {
			denom := (a.End - a.Start)
			if denom != 0 {
				v = (a.Now - a.Start) / denom
			} else {
				v = 1
			}
		} else {
			adj := seg.AdjacentSegment.Sector
			denom := (a.End - adj.TopZ.Now)
			if denom != 0 {
				v = (a.Now - adj.TopZ.Now) / denom
			} else {
				v = 1
			}
		}
		t[concepts.MatBasis2Y] = 1.0 - v
		t[concepts.MatTransY] = v
		if !seg.Surface.Transform.Attached {
			seg.Surface.Transform.Attach(vd.DB.Simulation)
		}
		seg.Surface.Transform.Now.From(&seg.Surface.Transform.Original)
		seg.Surface.Transform.Now.MulSelf(&t)

		t[concepts.MatBasis2Y] = v
		t[concepts.MatTransY] = 1.0 - v
		if !seg.HiSurface.Transform.Attached {
			seg.HiSurface.Transform.Attach(vd.DB.Simulation)
		}
		seg.HiSurface.Transform.Now.From(&seg.HiSurface.Transform.Original)
		seg.HiSurface.Transform.Now.MulSelf(&t)
	}
}

func (vd *VerticalDoorController) Always() {
	if vd.Sector.TopZ.Animation == nil {
		vd.setupAnimation()
	}

	a := vd.Sector.TopZ.Animation

	if a.Percent <= 0 {
		vd.State = behaviors.DoorStateOpen
		if vd.Intent == behaviors.DoorIntentOpen && vd.AutoClose {
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
		a.Active = true
	} else if vd.Intent == behaviors.DoorIntentClosed && vd.State != behaviors.DoorStateClosed {
		vd.State = behaviors.DoorStateClosing
		a.Reverse = false
		a.Active = true
	}

	vd.adjustTransforms()
}

func (vd *VerticalDoorController) Recalculate() {
	if vd.Sector != nil {
		vd.setupAnimation()
	}
}
