// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"log"
	"math"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

func BodySectorScript(scripts []*core.Script, body *core.Body, sector *core.Sector) {
	for _, script := range scripts {
		script.Vars["body"] = body
		script.Vars["sector"] = sector
		script.Act()
	}
}

func (bc *BodyController) Enter(eSector concepts.Entity) {
	if eSector == 0 {
		log.Printf("%v tried to enter nil sector", bc.Body.Entity)
		return
	}
	sector := core.SectorFromDb(bc.Body.DB, eSector)
	if sector == nil {
		log.Printf("%v tried to enter entity %v that's not a sector", bc.Body.Entity, eSector.String(bc.Body.DB))
		return
	}
	bc.Sector = sector
	bc.Sector.Bodies[bc.Body.Entity] = bc.Body
	bc.Body.SectorEntity = eSector

	if bc.Body.OnGround {
		floorZ, _ := bc.Sector.PointZ(concepts.DynamicNow, bc.Body.Pos.Now.To2D())
		p := &bc.Body.Pos.Now
		h := bc.Body.Size.Now[1] * 0.5
		if bc.Sector.FloorTarget == 0 && p[2]-h < floorZ {
			p[2] = floorZ + h
		}
	}
	BodySectorScript(bc.Sector.EnterScripts, bc.Body, bc.Sector)
}

func (bc *BodyController) Exit() {
	if bc.Sector == nil {
		log.Printf("%v tried to exit nil sector", bc.Body.Entity)
		return
	}
	BodySectorScript(bc.Sector.ExitScripts, bc.Body, bc.Sector)
	delete(bc.Sector.Bodies, bc.Body.Entity)
	bc.Body.SectorEntity = 0
}

func (bc *BodyController) PushBack(segment *core.SectorSegment) bool {
	d := segment.DistanceToPoint2(bc.pos2d)
	if d > bc.Body.Size.Now[0]*bc.Body.Size.Now[0]*0.25 {
		return false
	}
	side := segment.WhichSide(bc.pos2d)
	closest := segment.ClosestToPoint(bc.pos2d)
	delta := bc.pos2d.Sub(closest)
	d = delta.Length()
	if d > constants.IntersectEpsilon {
		delta[0] /= d
		delta[1] /= d
	} else {
		delta.From(&segment.Normal)
		delta[0] = -delta[0]
		delta[1] = -delta[1]
		d = 0
	}
	if side > 0 {
		delta.MulSelf(bc.Body.Size.Now[0]*0.5 - d)
	} else {
		log.Printf("PushBack: body is on the front-facing side of segment (%v units)\n", d)
		delta.MulSelf(-bc.Body.Size.Now[0]*0.5 - d)
	}
	// Apply the impulse at the same time
	bc.pos.To2D().AddSelf(delta)
	if d > 0 {
		bc.Body.Vel.Now.To2D().AddSelf(delta)
	}

	return true
}

func (bc *BodyController) findBodySector() {
	var closestSector *core.Sector

	for _, attachable := range bc.Body.DB.AllOfType(core.SectorComponentIndex) {
		sector := attachable.(*core.Sector)
		if sector.IsPointInside2D(bc.pos2d) {
			closestSector = sector
			break
		}
	}

	if closestSector == nil {
		p := bc.Body.Pos.Now.To2D()
		var closestSeg *core.SectorSegment
		closestDistance2 := math.MaxFloat64
		for _, attachable := range bc.Body.DB.AllOfType(core.SectorComponentIndex) {
			sector := attachable.(*core.Sector)
			for _, seg := range sector.Segments {
				dist2 := seg.DistanceToPoint2(p)
				if closestSector == nil || dist2 < closestDistance2 {
					closestDistance2 = dist2
					closestSector = sector
					closestSeg = seg
				}
			}
		}
		p = closestSeg.ClosestToPoint(p)
		bc.Body.Pos.Now[0] = p[0]
		bc.Body.Pos.Now[1] = p[1]
	}

	floorZ, ceilZ := closestSector.PointZ(concepts.DynamicNow, bc.pos2d)
	//log.Printf("F: %v, C:%v\n", floorZ, ceilZ)
	if bc.pos[2]-bc.halfHeight < floorZ || bc.pos[2]+bc.halfHeight > ceilZ {
		//log.Printf("Moved body %v to closest sector and adjusted Z from %v to %v", bc.Body.Entity, p[2], floorZ)
		bc.pos[2] = floorZ + bc.halfHeight
	}
	bc.Enter(closestSector.Entity)
	// Don't mark as collided because this is probably an initialization.
}

func (bc *BodyController) checkBodySegmentCollisions() {
	// See if we need to push back into the current sector.
	for _, segment := range bc.Sector.Segments {
		if segment.AdjacentSector != 0 && segment.PortalIsPassable {
			adj := core.SectorFromDb(bc.Sector.DB, segment.AdjacentSector)
			// We can still collide with a portal if the heights don't match.
			// If we're within limits, ignore the portal.
			floorZ, ceilZ := adj.PointZ(concepts.DynamicNow, bc.pos2d)
			if bc.pos[2]-bc.halfHeight+bc.Body.MountHeight >= floorZ &&
				bc.pos[2]+bc.halfHeight < ceilZ {
				continue
			}
		}
		if bc.PushBack(segment) {
			bc.collidedSegments = append(bc.collidedSegments, segment)
		}
	}
}

func (bc *BodyController) bodyTeleport() bool {
	for _, segment := range bc.Sector.Segments {
		if !segment.PortalTeleports || segment.AdjacentSegment == nil {
			continue
		}
		d := segment.DistanceToPoint2(bc.pos2d)
		if d > bc.Body.Size.Now[0]*bc.Body.Size.Now[0]*0.25 {
			continue
		}
		side := segment.WhichSide(bc.pos2d)
		if side < 0 {
			// Teleport position
			v := segment.PortalMatrix.Unproject(bc.pos2d)
			v = segment.AdjacentSegment.MirrorPortalMatrix.Project(v)
			bc.Body.Pos.Now[0] = v[0]
			bc.Body.Pos.Now[1] = v[1]
			// Teleport velocity
			trans := *bc.Body.Vel.Now.To2D()
			trans[0] += segment.A[0]
			trans[1] += segment.A[1]
			v = segment.PortalMatrix.Unproject(&trans)
			v = segment.AdjacentSegment.MirrorPortalMatrix.Project(v)
			bc.Body.Vel.Now[0] = v[0] - segment.AdjacentSegment.B[0]
			bc.Body.Vel.Now[1] = v[1] - segment.AdjacentSegment.B[1]
			// Calculate new facing angle
			bc.Body.Angle.Now = bc.Body.Angle.Now -
				math.Atan2(segment.Normal[1], segment.Normal[0])*concepts.Rad2deg +
				math.Atan2(segment.AdjacentSegment.Normal[1], segment.AdjacentSegment.Normal[0])*concepts.Rad2deg + 180
			for bc.Body.Angle.Now > 360 {
				bc.Body.Angle.Now -= 360
			}
			bc.Body.LastEnteredPortal = segment
			bc.Enter(segment.AdjacentSector)
			return true
		}
	}
	return false
}

func (bc *BodyController) bodyExitsSector() {
	// Exit the current sector.
	bc.Exit()

	if bc.bodyTeleport() {
		return
	}

	for _, segment := range bc.Sector.Segments {
		if segment.AdjacentSector == 0 {
			continue
		}
		adj := core.SectorFromDb(bc.Sector.DB, segment.AdjacentSector)
		floorZ, ceilZ := adj.PointZ(concepts.DynamicNow, bc.pos2d)
		if bc.pos[2]-bc.halfHeight+bc.Body.MountHeight >= floorZ &&
			bc.pos[2]+bc.halfHeight < ceilZ &&
			adj.IsPointInside2D(bc.pos2d) {
			// Hooray, we've handled case 5! Make sure Z is good.
			//log.Printf("Case 5! body = %v in sector %v, floor z = %v\n", p.StringHuman(), adj.Entity, floorZ)
			/*if p[2] < floorZ {
				e.Pos[2] = floorZ
				log.Println("Entity entering adjacent sector is lower than floorZ")
			}*/
			bc.Body.LastEnteredPortal = segment
			bc.Enter(segment.AdjacentSector)
			break
		}
	}

	if bc.Sector == nil {
		// Case 6! This is the worst.
		for _, component := range bc.Body.DB.AllOfType(core.SectorComponentIndex) {
			sector := component.(*core.Sector)
			floorZ, ceilZ := sector.PointZ(concepts.DynamicNow, bc.pos2d)
			if bc.pos[2]-bc.halfHeight+bc.Body.MountHeight >= floorZ &&
				bc.pos[2]+bc.halfHeight < ceilZ {
				for _, segment := range sector.Segments {
					if bc.PushBack(segment) {
						bc.collidedSegments = append(bc.collidedSegments, segment)
					}
				}
			}
		}
	}
}

func (bc *BodyController) bodyBounce(body *core.Body) {
	r_a := bc.Body.Size.Now[0] * 0.5
	r_b := body.Size.Now[0] * 0.5
	n := bc.pos.Sub(&body.Pos.Now).NormSelf()
	r_ap := n.Mul(-r_a)
	r_bp := n.Mul(r_b)
	// For now, assume no angular velocity. In the future, this may
	// change.
	vang_a1, vang_b1 := new(concepts.Vector3), new(concepts.Vector3)
	v_a1, v_b1 := &bc.Body.Vel.Now, &body.Vel.Now

	v_ap1 := v_a1.Add(vang_a1.Cross(r_ap))
	v_bp1 := v_b1.Add(vang_b1.Cross(r_bp))
	v_p1 := v_ap1.Sub(v_bp1)

	c_a := r_ap.Cross(n)
	c_b := r_bp.Cross(n)
	// Solid spheres
	i_a := bc.Body.Mass * r_a * r_a * 2.0 / 5.0
	i_b := body.Mass * r_b * r_b * 2.0 / 5.0
	e := 0.8
	if i_a > 0 && i_b > 0 {
		j := -(1.0 + e) * v_p1.Dot(n) / (1.0/bc.Body.Mass + 1.0/body.Mass + c_a.Dot(c_a)/i_a + c_b.Dot(c_b)/i_b)
		bc.Body.Vel.Now.AddSelf(n.Mul(j / bc.Body.Mass))
		body.Vel.Now.AddSelf(n.Mul(-j / body.Mass))
	} else if i_a > 0 {
		j := -(1.0 + e) * v_p1.Dot(n) / (1.0/bc.Body.Mass + c_a.Dot(c_a)/i_a)
		bc.Body.Vel.Now.AddSelf(n.Mul(j / bc.Body.Mass))
	} else if i_b > 0 {
		j := -(1.0 + e) * v_p1.Dot(n) / (1.0/body.Mass + c_b.Dot(c_b)/i_b)
		body.Vel.Now.AddSelf(n.Mul(-j / body.Mass))
	}
	//fmt.Printf("%v <-> %v = %v\n", bc.Body.String(), body.String(), diff)
}

func (bc *BodyController) bodyBodyCollide(sector *core.Sector) {
	for _, body := range sector.Bodies {
		if body == nil || body == bc.Body || !body.IsActive() {
			continue
		}
		// From https://www.myphysicslab.com/engine2D/collision-en.html
		d2 := bc.pos.Dist2(&body.Pos.Now)
		r_a := bc.Body.Size.Now[0] * 0.5
		r_b := body.Size.Now[0] * 0.5
		if d2 < (r_a+r_b)*(r_a+r_b) {
			item := behaviors.InventoryItemFromDb(body.DB, body.Entity)
			if item != nil && bc.Player != nil {
				itemClone := item.DB.LoadComponentWithoutAttaching(behaviors.InventoryItemComponentIndex, item.Serialize())
				bc.Player.Inventory = append(bc.Player.Inventory, itemClone.(*behaviors.InventoryItem))
				body.Active = false
				continue
			}
			bc.bodyBounce(body)
		}
	}
}

func (bc *BodyController) Collide() {
	// We've got several possibilities we need to handle:
	// 1.   The body has an un-initialized sector, but it's within a sector and doesn't need to be moved.
	// 2.   The body is outside of all sectors. Put it into the nearest sector.
	// 3.   The body is still in its current sector, but it's gotten too close to a wall and needs to be pushed back.
	// 4.   The body is outside of the current sector because it's gone past a wall and needs to be pushed back.
	// 5.   The body is outside of the current sector because it's gone through a portal and needs to change sectors.
	// 6.   The body is outside of the current sector because it's gone through a portal, but it winds up outside of
	//      any sectors and needs to be pushed back into a valid sector using any walls within bounds.
	// 7.   The body has collided with another body nearby. Both bodies need to
	//      be pushed apart.
	// 8.   No collision occured.

	// The central method here is to push back, but the wall that's doing the pushing requires some logic to get.

	// Do 10 collision iterations to avoid spending too much time here.
	for i := 0; i < 10; i++ {
		// Avoid GC thrash
		bc.collidedSegments = bc.collidedSegments[:0]
		// Cases 1 & 2.
		if bc.Sector == nil {
			bc.findBodySector()
		}

		// Case 3 & 4
		bc.checkBodySegmentCollisions()

		if !bc.Sector.IsPointInside2D(bc.pos2d) {
			// Cases 5 & 6
			bc.bodyExitsSector()
		}

		if bc.Sector != nil {
			//		bc.bodyBodyCollide(bc.Sector)
			for _, sector := range bc.Sector.PVS {
				bc.bodyBodyCollide(sector)
			}
		}

		if len(bc.collidedSegments) == 0 {
			if bc.Sector != nil {
				return
			} else {
				continue
			}
		}

		for _, seg := range bc.collidedSegments {
			BodySectorScript(seg.ContactScripts, bc.Body, bc.Sector)
		}

		switch bc.Body.CollisionResponse {
		case core.Stop:
			bc.Body.Vel.Now[0] = 0
			bc.Body.Vel.Now[1] = 0
		case core.Bounce:
			for _, segment := range bc.collidedSegments {
				n := segment.Normal.To3D(new(concepts.Vector3))
				bc.Body.Vel.Now.SubSelf(n.Mul(2 * bc.Body.Vel.Now.Dot(n)))
			}
		case core.Remove:
			bc.RemoveBody()
		}
	}
}
