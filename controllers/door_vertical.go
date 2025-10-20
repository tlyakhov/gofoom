package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

func (d *DoorController) setupVerticalDoorAnimation(refresh bool) {
	if d.Sector.Top.Z.Animation != nil && !refresh {
		return
	}
	a := d.Sector.Top.Z.NewAnimation()
	a.Start = d.Sector.Top.Z.Spawn
	if d.UseLimit {
		a.End = d.Limit
	} else {
		a.End = d.Sector.Bottom.Z.Spawn
	}
	a.Coordinates = dynamic.AnimationCoordinatesAbsolute
	a.Duration = d.Duration
	a.TweeningFunc = d.TweeningFunc
	a.Lifetime = dynamic.AnimationLifetimeOnce
}

func (d *DoorController) calculateVerticalDoorTransforms() {
	a := d.Sector.Top.Z.Animation

	if a.Now == a.Prev {
		return
	}

	t := concepts.Matrix2{}
	t.SetIdentity()
	var v float64
	for _, seg := range d.Sector.Segments {
		if seg.AdjacentSegment == nil {
			denom := (d.Sector.Bottom.Z.Spawn - d.Sector.Top.Z.Spawn)
			if denom != 0 {
				v = (a.Now - d.Sector.Top.Z.Spawn) / denom
			} else {
				v = 1
			}
		} else {
			adj := seg.AdjacentSegment.Sector
			denom := (d.Sector.Bottom.Z.Spawn - adj.Top.Z.Now)
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

func (d *DoorController) checkVerticalDoorState() {
	a := d.Sector.Top.Z.Animation
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
