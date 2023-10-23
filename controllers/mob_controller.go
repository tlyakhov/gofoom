package controllers

import (
	"fmt"
	"log"
	"math"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type MobController struct {
	concepts.BaseController
	Mob    *core.Mob
	Sector *core.Sector
}

func init() {
	concepts.DbTypes().RegisterController(MobController{})
}

func (mc *MobController) Target(target *concepts.EntityRef) bool {
	mc.TargetEntity = target
	mc.Mob = core.MobFromDb(target)
	mc.Sector = mc.Mob.Sector()
	return mc.Mob != nil && mc.Mob.Active
}

func (mc *MobController) PushBack(segment *core.Segment) bool {
	p2d := mc.Mob.Pos.Now.To2D()
	d := segment.DistanceToPoint2(p2d)
	if d > mc.Mob.BoundingRadius*mc.Mob.BoundingRadius {
		return false
	}
	side := segment.WhichSide(p2d)
	closest := segment.ClosestToPoint(p2d)
	delta := p2d.Sub(closest)
	d = delta.Length()
	delta.NormSelf()
	if side > 0 {
		delta.MulSelf(mc.Mob.BoundingRadius - d)
	} else {
		log.Printf("PushBack: mob is on the front-facing side of segment (%v units)\n", d)
		delta.MulSelf(-mc.Mob.BoundingRadius - d)
	}
	// Apply the impulse at the same time
	mc.Mob.Pos.Now.To2D().AddSelf(delta)
	mc.Mob.Vel.Now.To2D().AddSelf(delta)

	return true
}

func (mc *MobController) Collide() []*core.Segment {
	// We've got several possibilities we need to handle:
	// 1.   The mob is outside of all sectors. Put it into the nearest sector.
	// 2.   The mob has an un-initialized sector, but it's within a sector and doesn't need to be moved.
	// 3.   The mob is still in its current sector, but it's gotten too close to a wall and needs to be pushed back.
	// 4.   The mob is outside of the current sector because it's gone past a wall and needs to be pushed back.
	// 5.   The mob is outside of the current sector because it's gone through a portal and needs to change sectors.
	// 6.   The mob is outside of the current sector because it's gone through a portal, but it winds up outside of
	//      any sectors and needs to be pushed back into a valid sector using any walls within bounds.
	// 7.   No collision occured.

	// The central method here is to push back, but the wall that's doing the pushing requires some logic to get.

	// Assume we haven't collided.
	var collided []*core.Segment
	p := &mc.Mob.Pos.Now

	// Cases 1 & 2.
	if mc.Sector == nil {
		var closestSector *core.Sector
		closestDistance2 := math.MaxFloat64

		for _, component := range mc.Mob.DB.All(core.SectorComponentIndex) {
			sector := component.(*core.Sector)
			d2 := p.Dist2(&sector.Center)

			if closestSector == nil || d2 < closestDistance2 {
				closestDistance2 = d2
				closestSector = sector
			}
			if sector.IsPointInside2D(p.To2D()) {
				closestDistance2 = 0
				break
			}
		}

		if closestDistance2 > 0 {
			mc.Mob.Pos.Now = closestSector.Center
		}

		floorZ, ceilZ := closestSector.SlopedZNow(p.To2D())
		if p[2] < floorZ || p[2]+mc.Mob.Height > ceilZ {
			log.Printf("Moved mob %v to closest sector and adjusted Z from %v to %v", mc.Mob.Entity, p[2], floorZ)
			p[2] = floorZ
		}
		mc.Sector = closestSector
		mc.ControllerSet.Act(mc.TargetEntity, closestSector.EntityRef(), "Enter")
		// Don't mark as collided because this is probably an initialization.
	}

	sector := mc.Sector

	// Case 3 & 4
	// See if we need to push back into the current sector.
	for _, segment := range sector.Segments {
		if !segment.AdjacentSector.Nil() {
			adj := core.SectorFromDb(segment.AdjacentSector)
			// We can still collide with a portal if the heights don't match.
			// If we're within limits, ignore the portal.
			floorZ, ceilZ := adj.SlopedZNow(p.To2D())
			if p[2]+mc.Mob.MountHeight >= floorZ &&
				p[2]+mc.Mob.Height < ceilZ {
				continue
			}
		}
		if mc.PushBack(segment) {
			collided = append(collided, segment)
		}
	}

	ePosition2D := p.To2D()
	inSector := sector.IsPointInside2D(ePosition2D)
	if !inSector {
		// Cases 5 & 6

		// Exit the current sector.
		mc.ControllerSet.Act(mc.TargetEntity, mc.Mob.SectorEntityRef, "Exit")

		for _, segment := range sector.Segments {
			if segment.AdjacentSector.Nil() {
				continue
			}
			adj := core.SectorFromDb(segment.AdjacentSector)
			floorZ, ceilZ := adj.SlopedZNow(ePosition2D)
			if p[2]+mc.Mob.MountHeight >= floorZ &&
				p[2]+mc.Mob.Height < ceilZ &&
				adj.IsPointInside2D(ePosition2D) {
				// Hooray, we've handled case 5! Make sure Z is good.
				fmt.Printf("Case 5! mob = %v, floor z = %v\n", p, floorZ)
				if p[2] < floorZ {
					//e.Pos[2] = floorZ
					fmt.Println("goop2?")
				}
				mc.Mob.SectorEntityRef = segment.AdjacentSector
				mc.Sector = adj
				mc.ControllerSet.Act(mc.TargetEntity, mc.Mob.SectorEntityRef, "Enter")
				break
			}
		}

		if mc.Sector == nil {
			// Case 6! This is the worst.
			for _, component := range mc.Mob.DB.All(core.SectorComponentIndex) {
				sector := component.(*core.Sector)
				floorZ, ceilZ := sector.SlopedZNow(p.To2D())
				if p[2]+mc.Mob.MountHeight >= floorZ &&
					p[2]+mc.Mob.Height < ceilZ {
					for _, segment := range sector.Segments {
						if mc.PushBack(segment) {
							collided = append(collided, segment)
						}
					}
				}
			}
			//children := e.Collide() // RECURSIVE! Can be infinite, I suppose?
			//collided = append(collided, children...)
		}
	}

	if len(collided) > 0 {
		response := mc.Mob.CollisionResponse
		mc.ControllerSet.Act(mc.TargetEntity, mc.Mob.SectorEntityRef, "Contact")

		if response == core.Stop {
			mc.Mob.Vel.Now[0] = 0
			mc.Mob.Vel.Now[1] = 0
		} else if response == core.Bounce {
			for _, segment := range collided {
				n := segment.Normal.To3D(new(concepts.Vector3))
				mc.Mob.Vel.Now.SubSelf(n.Mul(2 * mc.Mob.Vel.Now.Dot(n)))
			}
		} else if response == core.Remove {
			mc.RemoveMob()
		}
	}

	mob := mc.Sector.Mobs[mc.Mob.Entity]
	if mc.Sector != nil && mob.Nil() {
		mc.Sector.Mobs[mc.Mob.Entity] = *mc.TargetEntity
	}

	return collided
}

func (mc *MobController) RemoveMob() {
	// TODO: poorly implemented
	if mc.Sector != nil {
		delete(mc.Sector.Mobs, mc.Mob.Entity)
		mc.Sector = nil
		mc.Mob.SectorEntityRef.Reset()
		//return
	}
	panic("MobController.RemoveMob is broken")
}

func (mc *MobController) ResetForce() {
	mc.Mob.Force[2] = 0.0
	mc.Mob.Force[1] = 0.0
	mc.Mob.Force[0] = 0.0
}

func (mc *MobController) Always() {
	if mc.Mob.Mass == 0 {
		// Reset force for next frame
		mc.ResetForce()
		return
	}
	// Our physics are impulse-based. We do semi-implicit Euler calculations
	// at each time step, and apply constraints (e.g. collision) directly to the velocities
	// f = ma
	// a = f/m
	// v = ∫a dt
	// p = ∫v dt
	mc.Mob.Vel.Now.AddSelf(mc.Mob.Force.Mul(constants.TimeStepS / mc.Mob.Mass))
	if mc.Mob.Vel.Now.Length2() > constants.VelocityEpsilon {
		speed := mc.Mob.Vel.Now.Length() * constants.TimeStepS
		steps := concepts.Min(concepts.Max(int(speed/constants.CollisionCheck), 1), 10)
		dt := constants.TimeStepS / float64(steps)
		for step := 0; step < steps; step++ {
			mc.Mob.Pos.Now.AddSelf(mc.Mob.Vel.Now.Mul(dt * constants.UnitsPerMeter))
			// Constraint impulses
			mc.Collide()
		}
	}
	// Reset force for next frame
	mc.ResetForce()
}

func (mc *MobController) Loaded() {
	mc.Collide()
}
