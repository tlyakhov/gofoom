// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
)

type DoorController struct {
	ecs.BaseController
	*behaviors.Door
	Sector *core.Sector

	autoProximity *behaviors.Proximity
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &DoorController{} }, 100)
}

func (d *DoorController) ComponentID() ecs.ComponentID {
	return behaviors.DoorCID
}

func (d *DoorController) Methods() ecs.ControllerMethod {
	return ecs.ControllerFrame | ecs.ControllerRecalculate
}

func (d *DoorController) Target(target ecs.Component, e ecs.Entity) bool {
	d.Entity = e
	d.Door = target.(*behaviors.Door)
	d.Sector = core.GetSector(d.Door.Entity)
	return d.Door.IsActive() && d.Sector != nil && d.Sector.IsActive()
}

func (d *DoorController) open() {
	switch d.Type {
	case behaviors.DoorTypeVertical:
		d.Sector.Top.Z.Animation.Reverse = true
		d.Sector.Top.Z.Animation.Active = true
	case behaviors.DoorTypeSwing:
		d.Sector.Transform.Animation.Reverse = true
		d.Sector.Transform.Animation.Active = true
	}
	d.Open.Vars["door"] = d.Door
	d.Open.Act()
}

func (d *DoorController) close() {
	switch d.Type {
	case behaviors.DoorTypeVertical:
		d.Sector.Top.Z.Animation.Reverse = false
		d.Sector.Top.Z.Animation.Active = true
	case behaviors.DoorTypeSwing:
		d.Sector.Transform.Animation.Reverse = false
		d.Sector.Transform.Animation.Active = true
	}
	d.Close.Vars["door"] = d.Door
	d.Close.Act()
}

func (d *DoorController) Frame() {
	switch d.Type {
	case behaviors.DoorTypeVertical:
		d.setupVerticalDoorAnimation(false)
	case behaviors.DoorTypeSwing:
		d.setupSwingDoorAnimation(false)
	}

	if d.Intent == behaviors.DoorIntentOpen && d.State != behaviors.DoorStateOpen && d.State != behaviors.DoorStateOpening {
		d.State = behaviors.DoorStateOpening
		d.open()
	} else if d.Intent == behaviors.DoorIntentClosed && d.State != behaviors.DoorStateClosed && d.State != behaviors.DoorStateClosing {
		d.State = behaviors.DoorStateClosing
		d.close()
	}

	switch d.Type {
	case behaviors.DoorTypeVertical:
		d.checkVerticalDoorState()
		d.calculateVerticalDoorTransforms()
	case behaviors.DoorTypeSwing:
		d.checkSwingDoorState()
	}
}

func (d *DoorController) Recalculate() {
	if d.Sector != nil {
		switch d.Type {
		case behaviors.DoorTypeVertical:
			d.setupVerticalDoorAnimation(true)
		case behaviors.DoorTypeSwing:
			d.setupSwingDoorAnimation(true)
		}
	}
	if !d.Open.IsEmpty() {
		d.Open.Params = []core.ScriptParam{{Name: "door", TypeName: "*behaviors.Door"}}
		d.Open.Compile()
	}
	if !d.Close.IsEmpty() {
		d.Close.Params = []core.ScriptParam{{Name: "door", TypeName: "*behaviors.Door"}}
		d.Close.Compile()
	}

	if d.AutoProximity {
		d.cacheAutoProximity()
		p := behaviors.GetProximity(d.Entity)
		if p != nil && p != d.autoProximity {
			ecs.DetachComponent(behaviors.ProximityCID, d.Entity)
			p = nil
		}
		if p == nil {
			var a ecs.Component = d.autoProximity
			ecs.Attach(behaviors.ProximityCID, d.Entity, &a)
		}
	}
}

func (d *DoorController) cacheAutoProximity() {
	if ecs.CachedGeneratedComponent(&d.autoProximity, "_DoorAutoProximity", behaviors.ProximityCID) {
		d.autoProximity.Hysteresis = 0
		d.autoProximity.IgnoreSectorTransform = true
		d.autoProximity.InRange.Code = `
			if body == nil { return }
   			m := core.GetMobile(body.Entity)
   			vd := behaviors.GetDoor(onEntity)
   			if m == nil || vd == nil { return }
   			if m.Mass > 10 { vd.Intent = behaviors.DoorIntentOpen }`

		ecs.ActAllControllersOneEntity(d.autoProximity.Entity, ecs.ControllerRecalculate)
	}
}
