package mob_controllers

import (
	"fmt"
	"log"
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/controllers/provide"
	"tlyakhov/gofoom/core"
)

type PhysicalMobController struct {
	*core.PhysicalMob
}

func NewPhysicalMobController(pe *core.PhysicalMob) *PhysicalMobController {
	return &PhysicalMobController{PhysicalMob: pe}
}

func (e *PhysicalMobController) PushBack(segment *core.Segment) bool {
	p2d := e.Pos.Now.To2D()
	d := segment.DistanceToPoint2(p2d)
	if d > e.BoundingRadius*e.BoundingRadius {
		return false
	}
	side := segment.WhichSide(p2d)
	closest := segment.ClosestToPoint(p2d)
	delta := p2d.Sub(closest)
	d = delta.Length()
	delta.NormSelf()
	if side > 0 {
		delta.MulSelf(e.BoundingRadius - d)
	} else {
		log.Printf("PushBack: mob is on the front-facing side of segment (%v units)\n", d)
		delta.MulSelf(-e.BoundingRadius - d)
	}
	// Apply the impulse at the same time
	e.Pos.Now.To2D().AddSelf(delta)
	e.Vel.Now.To2D().AddSelf(delta)

	return true
}

func (e *PhysicalMobController) Collide() []*core.Segment {
	if e.Map == nil {
		return nil
	}
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
	p := &e.Pos.Now
	model := e.Model.(core.AbstractMob)

	// Cases 1 & 2.
	if e.Sector == nil {
		var closestSector core.AbstractSector
		closestDistance2 := math.MaxFloat64

		for _, sector := range e.Map.Sectors {
			d2 := p.Dist2(&sector.Physical().Center)

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
			e.Pos.Now = closestSector.Physical().Center
		}

		floorZ, ceilZ := closestSector.Physical().SlopedZNow(p.To2D())
		if p[2] < floorZ || p[2]+e.Height > ceilZ {
			log.Printf("Moved mob %v to closest sector and adjusted Z from %v to %v", e.Name, p[2], floorZ)
			p[2] = floorZ
		}

		e.Sector = closestSector
		provide.Passer.For(closestSector).OnEnter(model)
		// Don't mark as collided because this is probably an initialization.
	}

	// Case 3 & 4
	// See if we need to push back into the current sector.
	for _, segment := range e.Sector.Physical().Segments {
		if segment.AdjacentSector != nil {
			adj := segment.AdjacentSector.Physical()
			// We can still collide with a portal if the heights don't match.
			// If we're within limits, ignore the portal.
			floorZ, ceilZ := adj.SlopedZNow(p.To2D())
			if p[2]+e.MountHeight >= floorZ &&
				p[2]+e.Height < ceilZ {
				continue
			}
		}
		if e.PushBack(segment) {
			collided = append(collided, segment)
		}
	}

	ePosition2D := p.To2D()
	inSector := e.Sector.IsPointInside2D(ePosition2D)
	if !inSector {
		// Cases 5 & 6

		// Exit the current sector.
		sector := e.Sector
		provide.Passer.For(e.Sector).OnExit(model)
		e.Sector = nil

		for _, segment := range sector.Physical().Segments {
			if segment.AdjacentSector == nil {
				continue
			}
			adj := segment.AdjacentSector.Physical()
			floorZ, ceilZ := adj.SlopedZNow(ePosition2D)
			if p[2]+e.MountHeight >= floorZ &&
				p[2]+e.Height < ceilZ &&
				adj.IsPointInside2D(ePosition2D) {
				// Hooray, we've handled case 5! Make sure Z is good.
				fmt.Printf("Case 5! mob = %v, floor z = %v\n", p, floorZ)
				if p[2] < floorZ {
					//e.Pos[2] = floorZ
					fmt.Println("goop2?")
				}
				e.Sector = segment.AdjacentSector
				provide.Passer.For(e.Sector).OnEnter(model)
				break
			}
		}

		if e.Sector == nil {
			// Case 6! This is the worst.
			for _, sector := range e.Map.Sectors {
				phys := sector.Physical()
				floorZ, ceilZ := phys.SlopedZNow(p.To2D())
				if p[2]+e.MountHeight >= floorZ &&
					p[2]+e.Height < ceilZ {
					for _, segment := range phys.Segments {
						if e.PushBack(segment) {
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
		response := e.CollisionResponse

		if e.CRCallback != nil {
			response = e.CRCallback()
		}

		if response == core.Stop {
			e.Vel.Now[0] = 0
			e.Vel.Now[1] = 0
		} else if response == core.Bounce {
			for _, segment := range collided {
				n := segment.Normal.To3D(new(concepts.Vector3))
				e.Vel.Now.SubSelf(n.Mul(2 * e.Vel.Now.Dot(n)))
			}
		} else if response == core.Remove {
			e.Remove()
		}
	}

	if e.Sector != nil && e.Sector.Physical().Mobs[e.Name] == nil {
		e.Sector.Physical().Mobs[e.Name] = model
	}

	return collided
}

func (e *PhysicalMobController) Remove() {
	if e.Sector != nil {
		delete(e.Sector.Physical().Mobs, e.Name)
		e.Sector = nil
		return
	}

	for _, sector := range e.Map.Sectors {
		delete(sector.Physical().Mobs, e.Name)
	}
}

func (e *PhysicalMobController) Frame() {
	if !e.Active {
		return
	}

	if e.Mass != 0 {
		// Our physics are impulse-based. We do semi-implicit Euler calculations
		// at each time step, and apply constraints (e.g. collision) directly to the velocities
		// f = ma
		// a = f/m
		// v = ∫a dt
		// p = ∫v dt
		e.Vel.Now.AddSelf(e.Force.Mul(constants.TimeStepS / e.Mass))
		if e.Vel.Now.Length2() > constants.VelocityEpsilon {
			speed := e.Vel.Now.Length() * constants.TimeStepS
			steps := concepts.Min(concepts.Max(int(speed/constants.CollisionCheck), 1), 10)
			dt := constants.TimeStepS / float64(steps)
			for step := 0; step < steps; step++ {
				e.Pos.Now.AddSelf(e.Vel.Now.Mul(dt * constants.UnitsPerMeter))
				// Constraint impulses
				e.Collide()
				/*
					collSegments := e.Collide()
					if collSegments != nil {
						break
					}*/
			}
		}
	}
	// Reset force for next frame
	e.Force[2] = 0.0
	e.Force[1] = 0.0
	e.Force[0] = 0.0
}
