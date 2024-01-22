package controllers

import (
	"log"
	"math"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type BodyController struct {
	concepts.BaseController
	Body   *core.Body
	Sector *core.Sector
}

func init() {
	concepts.DbTypes().RegisterController(&BodyController{})
}

func (bc *BodyController) ComponentIndex() int {
	return core.BodyComponentIndex
}

func (bc *BodyController) Priority() int {
	return 80
}

func (bc *BodyController) Methods() concepts.ControllerMethod {
	return concepts.ControllerAlways |
		concepts.ControllerRecalculate |
		concepts.ControllerLoaded
}

func (bc *BodyController) Target(target concepts.Attachable) bool {
	bc.Body = target.(*core.Body)
	if !bc.Body.IsActive() {
		return false
	}
	bc.Sector = bc.Body.Sector()
	return true
}

func (bc *BodyController) PushBack(segment *core.Segment) bool {
	p2d := bc.Body.Pos.Now.To2D()
	d := segment.DistanceToPoint2(p2d)
	if d > bc.Body.Size.Now[0]*bc.Body.Size.Now[0]*0.25 {
		return false
	}
	side := segment.WhichSide(p2d)
	closest := segment.ClosestToPoint(p2d)
	delta := p2d.Sub(closest)
	d = delta.Length()
	delta.NormSelf()
	if side > 0 {
		delta.MulSelf(bc.Body.Size.Now[0]*0.5 - d)
	} else {
		log.Printf("PushBack: body is on the front-facing side of segment (%v units)\n", d)
		delta.MulSelf(-bc.Body.Size.Now[0]*0.5 - d)
	}
	// Apply the impulse at the same time
	bc.Body.Pos.Now.To2D().AddSelf(delta)
	bc.Body.Vel.Now.To2D().AddSelf(delta)

	return true
}

func (bc *BodyController) Collide() []*core.Segment {
	// We've got several possibilities we need to handle:
	// 1.   The body is outside of all sectors. Put it into the nearest sector.
	// 2.   The body has an un-initialized sector, but it's within a sector and doesn't need to be moved.
	// 3.   The body is still in its current sector, but it's gotten too close to a wall and needs to be pushed back.
	// 4.   The body is outside of the current sector because it's gone past a wall and needs to be pushed back.
	// 5.   The body is outside of the current sector because it's gone through a portal and needs to change sectors.
	// 6.   The body is outside of the current sector because it's gone through a portal, but it winds up outside of
	//      any sectors and needs to be pushed back into a valid sector using any walls within bounds.
	// 7.   No collision occured.

	// The central method here is to push back, but the wall that's doing the pushing requires some logic to get.

	// Assume we haven't collided.
	var collided []*core.Segment
	p := &bc.Body.Pos.Now

	// Cases 1 & 2.
	if bc.Sector == nil {
		var closestSector *core.Sector
		closestDistance2 := math.MaxFloat64

		for _, isector := range bc.EntityComponentDB.All(core.SectorComponentIndex) {
			sector := isector.(*core.Sector)
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
			bc.Body.Pos.Now = closestSector.Center
		}

		floorZ, ceilZ := closestSector.SlopedZNow(p.To2D())
		//log.Printf("F: %v, C:%v\n", floorZ, ceilZ)
		if p[2] < floorZ || p[2]+bc.Body.Size.Now[1] > ceilZ {
			//log.Printf("Moved body %v to closest sector and adjusted Z from %v to %v", bc.Body.Entity, p[2], floorZ)
			p[2] = floorZ
		}
		bc.Enter(closestSector.Ref())
		// Don't mark as collided because this is probably an initialization.
	}

	sector := bc.Sector

	// Case 3 & 4
	// See if we need to push back into the current sector.
	for _, segment := range sector.Segments {
		if !segment.AdjacentSector.Nil() && segment.PortalIsPassable {
			adj := core.SectorFromDb(segment.AdjacentSector)
			// We can still collide with a portal if the heights don't match.
			// If we're within limits, ignore the portal.
			floorZ, ceilZ := adj.SlopedZNow(p.To2D())
			if p[2]+bc.Body.MountHeight >= floorZ &&
				p[2]+bc.Body.Size.Now[1] < ceilZ {
				continue
			}
		}
		if bc.PushBack(segment) {
			collided = append(collided, segment)
		}
	}

	ePosition2D := p.To2D()
	inSector := sector.IsPointInside2D(ePosition2D)
	if !inSector {
		// Cases 5 & 6

		// Exit the current sector.
		bc.Exit()

		for _, segment := range sector.Segments {
			if segment.AdjacentSector.Nil() {
				continue
			}
			adj := core.SectorFromDb(segment.AdjacentSector)
			floorZ, ceilZ := adj.SlopedZNow(ePosition2D)
			if p[2]+bc.Body.MountHeight >= floorZ &&
				p[2]+bc.Body.Size.Now[1] < ceilZ &&
				adj.IsPointInside2D(ePosition2D) {
				// Hooray, we've handled case 5! Make sure Z is good.
				//log.Printf("Case 5! body = %v in sector %v, floor z = %v\n", p.StringHuman(), adj.Entity, floorZ)
				/*if p[2] < floorZ {
					e.Pos[2] = floorZ
					log.Println("Entity entering adjacent sector is lower than floorZ")
				}*/
				bc.Enter(segment.AdjacentSector)
				break
			}
		}

		if bc.Sector == nil {
			// Case 6! This is the worst.
			for _, component := range bc.Body.DB.All(core.SectorComponentIndex) {
				sector := component.(*core.Sector)
				floorZ, ceilZ := sector.SlopedZNow(p.To2D())
				if p[2]+bc.Body.MountHeight >= floorZ &&
					p[2]+bc.Body.Size.Now[1] < ceilZ {
					for _, segment := range sector.Segments {
						if bc.PushBack(segment) {
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
		for _, seg := range collided {
			BodySectorScript(seg.ContactScripts, bc.Body.EntityRef, bc.Sector.EntityRef)
		}

		switch bc.Body.CollisionResponse {
		case core.Stop:
			bc.Body.Vel.Now[0] = 0
			bc.Body.Vel.Now[1] = 0
		case core.Bounce:
			for _, segment := range collided {
				n := segment.Normal.To3D(new(concepts.Vector3))
				bc.Body.Vel.Now.SubSelf(n.Mul(2 * bc.Body.Vel.Now.Dot(n)))
			}
		case core.Remove:
			bc.RemoveBody()
		}
	}

	return collided
}

func (bc *BodyController) RemoveBody() {
	// TODO: poorly implemented
	if bc.Sector != nil {
		delete(bc.Sector.Bodies, bc.Body.Entity)
		bc.Sector = nil
		bc.Body.SectorEntityRef = nil
		//return
	}
	panic("BodyController.RemoveBody is broken")
}

func (bc *BodyController) ResetForce() {
	bc.Body.Force[2] = 0.0
	bc.Body.Force[1] = 0.0
	bc.Body.Force[0] = 0.0
}

func (bc *BodyController) Always() {
	if bc.Body.Mass == 0 {
		// Reset force for next frame
		bc.ResetForce()
		return
	}
	if bc.Sector != nil {
		bc.Physics()
	}
	// Our physics are impulse-based. We do semi-implicit Euler calculations
	// at each time step, and apply constraints (e.g. collision) directly to the velocities
	// f = ma
	// a = f/m
	// v = ∫a dt
	// p = ∫v dt
	bc.Body.Vel.Now.AddSelf(bc.Body.Force.Mul(constants.TimeStepS / bc.Body.Mass))
	if bc.Body.Vel.Now.Length2() > constants.VelocityEpsilon {
		speed := bc.Body.Vel.Now.Length() * constants.TimeStepS
		steps := concepts.Min(concepts.Max(int(speed/constants.CollisionCheck), 1), 10)
		dt := constants.TimeStepS / float64(steps)
		for step := 0; step < steps; step++ {
			bc.Body.Pos.Now.AddSelf(bc.Body.Vel.Now.Mul(dt * constants.UnitsPerMeter))
			// Constraint impulses
			bc.Collide()
		}
	}
	// Reset force for next frame
	bc.ResetForce()
}

func (bc *BodyController) Recalculate() {
	bc.Collide()
}

func (bc *BodyController) Loaded() {
	bc.Collide()
}
