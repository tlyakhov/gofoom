package entity

import (
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/logic/provide"
)

type PhysicalEntityService struct {
	*core.PhysicalEntity
}

func NewPhysicalEntityService(e *core.PhysicalEntity) *PhysicalEntityService {
	return &PhysicalEntityService{PhysicalEntity: e}
}

func (e *PhysicalEntityService) PushBack(segment *core.Segment) bool {
	p2d := e.Pos.To2D()
	d := segment.DistanceToPoint2(p2d)
	if d > e.BoundingRadius*e.BoundingRadius {
		return false
	}
	side := segment.WhichSide(p2d)
	closest := segment.ClosestToPoint(p2d)
	v := e.Pos.Sub(closest.To3D())
	v.Z = 0
	d = v.Length()
	v = v.Norm()
	if side > 0 {
		v = v.Mul(e.BoundingRadius - d)
	} else {
		v = v.Mul(-e.BoundingRadius - d)
	}
	e.Pos = e.Pos.Add(v)

	return true
}

func (e *PhysicalEntityService) Collide() []*core.Segment {
	if e.Map == nil {
		return nil
	}
	// We've got several possibilities we need to handle:
	// 1.   The entity is outside of all sectors. Put it into the nearest sector.
	// 2.   The entity has an un-initialized sector, but it's within a sector and doesn't need to be moved.
	// 3.   The entity is still in its current sector, but it's gotten too close to a wall and needs to be pushed back.
	// 4.   The entity is outside of the current sector because it's gone past a wall and needs to be pushed back.
	// 5.   The entity is outside of the current sector because it's gone through a portal and needs to change sectors.
	// 6.   The entity is outside of the current sector because it's gone through a portal, but it winds up outside of
	//      any sectors and needs to be pushed back into a valid sector using any walls within bounds.
	// 7.   No collision occured.

	// The central method here is to push back, but the wall that's doing the pushing requires some logic to get.

	// Assume we haven't collided.
	var collided []*core.Segment

	// Cases 1 & 2.
	if e.Sector == nil {
		var closestSector core.AbstractSector
		closestDistance2 := math.MaxFloat64

		for _, sector := range e.Map.Sectors {
			d2 := e.Pos.Dist2(sector.Physical().Center)

			if closestSector == nil || d2 < closestDistance2 {
				closestDistance2 = d2
				closestSector = sector
			}
		}

		if !closestSector.IsPointInside2D(e.Pos.To2D()) {
			e.Pos = closestSector.Physical().Center
		} else if e.Pos.Z < closestSector.Physical().BottomZ || e.Pos.Z+e.Height > closestSector.Physical().TopZ {
			e.Pos.Z = closestSector.Physical().Center.Z
		}

		e.Sector = closestSector
		closestSector.Physical().Entities[e.ID] = e.PhysicalEntity
		provide.Passer.For(closestSector).OnEnter(e.PhysicalEntity)
		// Don't mark as collided because this is probably an initialization.
	}

	// Case 3 & 4
	// See if we need to push back into the current sector.
	for _, segment := range e.Sector.Physical().Segments {
		if segment.AdjacentSector != nil {
			adj := segment.AdjacentSector.Physical()
			// We can still collide with a portal if the heights don't match.
			// If we're within limits, ignore the portal.
			if e.Pos.Z+e.MountHeight >= adj.BottomZ &&
				e.Pos.Z+e.Height < adj.TopZ {
				continue
			}
		}
		if e.PushBack(segment) {
			collided = append(collided, segment)
		}
	}

	ePosition2D := e.Pos.To2D()
	inSector := e.Sector.IsPointInside2D(ePosition2D)
	if !inSector {
		// Cases 5 & 6

		// Exit the current sector.
		provide.Passer.For(e.Sector).OnExit(e.Physical())
		delete(e.Sector.Physical().Entities, e.ID)
		e.Sector = nil

		for _, item := range e.Map.Sectors {
			sector := item.Physical()
			if e.Pos.Z+e.MountHeight >= sector.BottomZ &&
				e.Pos.Z+e.Height < sector.TopZ &&
				sector.IsPointInside2D(ePosition2D) {
				// Hooray, we've handled case 5! Make sure Z is good.
				if e.Pos.Z < sector.BottomZ {
					e.Pos.Z = sector.BottomZ
				}
				e.Sector = item
				e.Sector.Physical().Entities[e.ID] = e
				provide.Passer.For(e.Sector).OnEnter(e)
				break
			}
		}

		if e.Sector == nil {
			// Case 6! This is the worst.
			for _, item := range e.Map.Sectors {
				sector := item.Physical()
				if e.Pos.Z+e.MountHeight >= sector.BottomZ &&
					e.Pos.Z+e.Height < sector.TopZ {

					for _, segment := range sector.Segments {
						if e.PushBack(segment) {
							collided = append(collided, segment)
						}
					}
				}
			}
			children := e.Collide() // RECURSIVE! Can be infinite, I suppose?
			collided = append(collided, children...)
		}
	}

	if len(collided) > 0 {
		response := e.CollisionResponse

		if e.CRCallback != nil {
			response = e.CRCallback()
		}

		if response == core.Stop {
			e.Vel.X = 0
			e.Vel.Y = 0
		} else if response == core.Bounce {
			for _, segment := range collided {
				e.Vel = e.Vel.Sub(segment.Normal.To3D().Mul(2 * e.Vel.Dot(segment.Normal.To3D())))
			}
		} else if response == core.Remove {
			e.Remove()
		}
	}

	if e.Sector != nil {
		e.Sector.Physical().Entities[e.ID] = e
	}

	return collided
}

func (e *PhysicalEntityService) Remove() {
	if e.Sector != nil {
		delete(e.Sector.Physical().Entities, e.ID)
		e.Sector = nil
		return
	}

	for _, sector := range e.Map.Sectors {
		delete(sector.Physical().Entities, e.ID)
	}
}

func (e *PhysicalEntityService) Frame(lastFrameTime float64) {
	if !e.Active {
		return
	}

	frameScale := lastFrameTime / 10.0

	if math.Abs(e.Vel.X) > constants.VelocityEpsilon ||
		math.Abs(e.Vel.Y) > constants.VelocityEpsilon ||
		math.Abs(e.Vel.Z) > constants.VelocityEpsilon {
		speed := e.Vel.Length() * frameScale
		steps := concepts.Max(int(speed/constants.CollisionCheck), 1)
		for step := 0; step < steps; step++ {
			e.Pos = e.Pos.Add(e.Vel.Mul(frameScale / float64(steps)))

			collSegments := e.Collide()
			if collSegments != nil {
				//break
			}
		}
	}
}
