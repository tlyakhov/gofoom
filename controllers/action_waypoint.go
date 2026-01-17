package controllers

import (
	"math"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
)

func (ac *ActionController) Waypoint(waypoint *behaviors.ActionWaypoint) bool {
	state := ac.timedAction(&waypoint.ActionTimed)

	switch state {
	case timedDelayed:
		return false
	}

	pos := ac.pos(true)
	target := &waypoint.P

	// TODO: Be more clever about when we refresh this
	pathStale := ecs.Simulation.SimTimestamp-ac.State.LastPathGenerated > concepts.MillisToNanos(1000)
	if waypoint.UsePathFinder && (ac.State.Path == nil || ac.State.LastPathGenerated == 0 || pathStale) {
		pf := PathFinder{
			Start:  pos,
			End:    target,
			Radius: ac.Body.Size.Now[0] * 0.5,
			Step:   10,
		}
		if ac.Mobile != nil {
			pf.MountHeight = ac.Mobile.MountHeight
		}
		ac.State.Path = pf.ShortestPath()
		ac.State.LastPathGenerated = ecs.Simulation.SimTimestamp
	}

	if len(ac.State.Path) > 0 {
		// Need to figure out which point to use as a target
		closestIndex := 0
		closestDist := math.Inf(1)
		for i, p := range ac.State.Path {
			dist := p.DistSq(pos)
			if dist <= closestDist {
				closestDist = dist
				closestIndex = i
			}
		}
		if closestIndex < len(ac.State.Path)-1 {
			closestIndex++
		}
		target = &ac.State.Path[closestIndex]
	}

	// Have we reached the target?
	d := ac.Speed * constants.TimeStepS
	if pos.To2D().DistSq(target.To2D()) < d*d {
		ac.State.Path = nil
		return true
	}

	force := target.Sub(pos)
	if ac.NoZ || ac.Sector != nil {
		force[2] = 0
		if ac.Body != nil && ac.Body.Pos.Procedural {
			ac.Body.Pos.Input[2] = ac.Body.Pos.Now[2]
		}
	}
	dist := force.Length()
	if dist > 0 {
		angle := math.Atan2(force[1], force[0]) * concepts.Rad2deg
		speed := ac.Speed / dist
		force.MulSelf(speed)
		if ac.FaceNextWaypoint {
			switch {
			case ac.Sector != nil && ac.Sector.Transform.Procedural:
				ac.Sector.Transform.Input.SetRotation(angle)
			case ac.Sector != nil:
				ac.Sector.Transform.Now.SetRotation(angle)
			case ac.Body.Angle.Procedural:
				ac.Body.Angle.Input = angle
			default:
				ac.Body.Angle.Now = angle
			}
		}
	}
	switch {
	case ac.Sector != nil && ac.Sector.Transform.Procedural:
		ac.Sector.Transform.Input.TranslateSelf(force.To2D().MulSelf(constants.TimeStepS))
	case ac.Sector != nil:
		ac.Sector.Transform.Now.TranslateSelf(force.To2D().MulSelf(constants.TimeStepS))
	case ac.Mobile != nil:
		// TODO: Is this a hack?
		if !ac.Body.OnGround {
			force.MulSelf(0.1)
		}
		ac.Mobile.Force.AddSelf(force)
	case ac.Body.Pos.Procedural:
		ac.Body.Pos.Input.AddSelf(force.MulSelf(constants.TimeStepS))
		var bc BodyController
		if bc.Target(ac.Body, ac.Entity) {
			bc.findBodySector()
		}
	default:
		ac.Body.Pos.Now.AddSelf(force.MulSelf(constants.TimeStepS))
		var bc BodyController
		if bc.Target(ac.Body, ac.Entity) {
			bc.findBodySector()
		}
	}

	return false
}
