package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/dynamic"
)

func (d *DoorController) setupSwingDoorAnimation(refresh bool) {
	if d.Sector.Transform.Animation != nil && !refresh {
		return
	}
	a := d.Sector.Transform.NewAnimation()
	// Start is open, End is closed
	a.End = d.Sector.Transform.Spawn
	if d.UseLimit {
		a.Start = *d.Sector.Transform.Spawn.Rotate(d.Limit)
	} else {
		a.Start = *d.Sector.Transform.Spawn.Rotate(90)
	}
	a.Coordinates = dynamic.AnimationCoordinatesAbsolute
	a.Duration = d.Duration
	a.TweeningFunc = d.TweeningFunc
	a.Lifetime = dynamic.AnimationLifetimeOnce
}

func (d *DoorController) checkSwingDoorState() {
	a := d.Sector.Transform.Animation
	if a.Percent <= 0 {
		d.State = behaviors.DoorStateOpen
		if d.Intent == behaviors.DoorIntentOpen && d.AutoClose {
			d.Intent = behaviors.DoorIntentClosed
		}
	}
	if a.Percent >= 1 {
		d.State = behaviors.DoorStateClosed
		if d.Intent == behaviors.DoorIntentClosed {
			d.Intent = behaviors.DoorIntentReset
		}
	}
}
