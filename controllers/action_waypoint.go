package controllers

import (
	"math"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/pathfinding"
)

func (ac *ActionController) pathSectorValid(from, to *core.Sector, p *concepts.Vector2) bool {
	// Maybe this or previous sector is a fromDoor?
	fromDoor := behaviors.GetDoor(from.Entity)
	toDoor := behaviors.GetDoor(to.Entity)
	if fromDoor != nil || toDoor != nil {
		// TODO: Check for doors that NPCs can't walk through.
		return true
	}

	// If it's not a door, check that the sector height is mountable and isn't too narrow
	fz, cz := from.ZAt(p)
	afz, acz := to.ZAt(p)
	if ac.Mobile != nil && afz-fz > ac.Mobile.MountHeight {
		return false
	}
	if min(cz, acz)-max(fz, afz) < ac.Body.Size.Now[0]*0.5 {
		return false
	}
	return true
}

func (ac *ActionController) repath(start, end concepts.Vector3) {
	ac.State.Finder = &pathfinding.Finder{
		Start:       &start,
		StartSector: ac.startSector(),
		End:         &end,
		Radius:      ac.Body.Size.Now[0] * 0.5,
		Step:        10,
		SectorValid: ac.pathSectorValid,
	}
	ac.State.Path = ac.State.Finder.ShortestPath()
	ac.State.LastPathGenerated = ecs.Simulation.SimTimestamp
}

func (ac *ActionController) pathTarget(pos *concepts.Vector3, target *concepts.Vector3) *concepts.Vector3 {
	if len(ac.State.Path) == 0 {
		return target
	}
	// Need to figure out which point to use as a target
	targetIndex := len(ac.State.Path) - 1
	lastDistSq := math.Inf(1)
	for i, p := range ac.State.Path {
		dist := p.DistSq(pos)
		if dist > lastDistSq {
			targetIndex = i
			break
		}
		lastDistSq = dist
	}
	return &ac.State.Path[targetIndex]
}

func (ac *ActionController) Waypoint(waypoint *behaviors.ActionWaypoint) bool {
	state := ac.timedAction(&waypoint.ActionTimed)

	switch state {
	case timedDelayed:
		return false
	}

	pos := ac.pos(true)
	target := &waypoint.P

	// TODO: Be more clever about when we refresh this
	pathStale := ecs.Simulation.SimTimestamp-ac.State.LastPathGenerated > concepts.MillisToNanos(3000)
	if waypoint.UsePathFinder && (ac.State.Finder == nil || ac.State.Path == nil || !ac.State.Finder.End.EqualEpsilon(target) || pathStale) {
		ac.repath(*pos, *target)
	}

	target = ac.pathTarget(pos, target)

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
			force.MulSelf(0.01)
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
