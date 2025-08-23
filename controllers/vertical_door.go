// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type VerticalDoorController struct {
	ecs.BaseController
	*behaviors.VerticalDoor
	Sector *core.Sector
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &VerticalDoorController{} }, 100)
}

func (vd *VerticalDoorController) ComponentID() ecs.ComponentID {
	return behaviors.VerticalDoorCID
}

func (vd *VerticalDoorController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways | ecs.ControllerRecalculate
}

func (vd *VerticalDoorController) Target(target ecs.Attachable, e ecs.Entity) bool {
	vd.Entity = e
	vd.VerticalDoor = target.(*behaviors.VerticalDoor)
	vd.Sector = core.GetSector(vd.VerticalDoor.Entity)
	return vd.VerticalDoor.IsActive() && vd.Sector != nil && vd.Sector.IsActive()
}

func (vd *VerticalDoorController) setupAnimation() {
	a := vd.Sector.Top.Z.NewAnimation()
	//a.OnAttach()
	a.Construct(nil)
	a.Start = vd.Sector.Top.Z.Spawn
	a.End = vd.Sector.Bottom.Z.Spawn
	a.Coordinates = dynamic.AnimationCoordinatesAbsolute
	a.Duration = vd.Duration
	a.TweeningFunc = vd.TweeningFunc
	a.Lifetime = dynamic.AnimationLifetimeOnce
}

func (vd *VerticalDoorController) adjustTransforms() {
	a := vd.Sector.Top.Z.Animation

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
			denom := (a.End - adj.Top.Z.Now)
			if denom != 0 {
				v = (a.Now - adj.Top.Z.Now) / denom
			} else {
				v = 1
			}
		}
		t[concepts.MatBasis2Y] = 1.0 - v
		t[concepts.MatTransY] = v
		if !seg.Surface.Transform.Attached {
			seg.Surface.Transform.Attach(ecs.Simulation)
		}
		seg.Surface.Transform.Now.From(&seg.Surface.Transform.Spawn)
		seg.Surface.Transform.Now.MulSelf(&t)

		t[concepts.MatBasis2Y] = v
		t[concepts.MatTransY] = 1.0 - v
		if !seg.HiSurface.Transform.Attached {
			seg.HiSurface.Transform.Attach(ecs.Simulation)
		}
		seg.HiSurface.Transform.Now.From(&seg.HiSurface.Transform.Spawn)
		seg.HiSurface.Transform.Now.MulSelf(&t)
	}
}

func (vd *VerticalDoorController) Always() {
	if vd.Sector.Top.Z.Animation == nil {
		vd.setupAnimation()
	}

	a := vd.Sector.Top.Z.Animation

	if vd.Intent == behaviors.DoorIntentOpen && vd.State != behaviors.DoorStateOpen && vd.State != behaviors.DoorStateOpening {
		vd.State = behaviors.DoorStateOpening
		a.Reverse = true
		a.Active = true
		vd.Open.Vars["door"] = vd.VerticalDoor
		vd.Open.Act()
	} else if vd.Intent == behaviors.DoorIntentClosed && vd.State != behaviors.DoorStateClosed && vd.State != behaviors.DoorStateClosing {
		vd.State = behaviors.DoorStateClosing
		a.Reverse = false
		a.Active = true
		vd.Close.Vars["door"] = vd.VerticalDoor
		vd.Close.Act()
	}

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

	vd.adjustTransforms()
}

func (vd *VerticalDoorController) Recalculate() {
	if vd.Sector != nil {
		vd.setupAnimation()
	}
	if !vd.Open.IsEmpty() {
		vd.Open.Params = []core.ScriptParam{{Name: "door", TypeName: "*behaviors.VerticalDoor"}}
		vd.Open.Compile()
	}
	if !vd.Close.IsEmpty() {
		vd.Close.Params = []core.ScriptParam{{Name: "door", TypeName: "*behaviors.VerticalDoor"}}
		vd.Close.Compile()
	}
}
