package controllers

import (
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type SectorMobController struct {
	concepts.BaseController
	*core.Sector
	Mob *core.Mob
}

func init() {
	concepts.DbTypes().RegisterController(SectorMobController{})
}

func (s *SectorMobController) Source(er *concepts.EntityRef) bool {
	s.SourceEntity = er
	s.Mob = core.MobFromDb(er)
	return s.Mob != nil && s.Mob.Active
}

func (s *SectorMobController) Target(target *concepts.EntityRef) bool {
	s.TargetEntity = target
	s.Sector = core.SectorFromDb(target)
	return s.Sector != nil && s.Sector.Active
}

func (s *SectorMobController) Enter() {
	s.Sector.Mobs[s.SourceEntity.Entity] = *s.SourceEntity

	if s.Mob.OnGround {
		floorZ, _ := s.Sector.SlopedZNow(s.Mob.Pos.Now.To2D())
		p := &s.Mob.Pos.Now
		if s.FloorTarget.Nil() && p[2] < floorZ {
			p[2] = floorZ
		}
	}
}

func (s *SectorMobController) Exit() {
	delete(s.Mobs, s.SourceEntity.Entity)
}

func (s *SectorMobController) Containment() {
	f := &s.Mob.Force
	if s.Mob.Mass > 0 {
		// Weight = g*m
		f[2] -= constants.Gravity * s.Mob.Mass
		v := &s.Mob.Vel.Now
		if !v.Zero() {
			// Air drag
			r := s.Mob.BoundingRadius * constants.MetersPerUnit
			crossSectionArea := math.Pi * r * r
			drag := concepts.Vector3{v[0], v[1], v[2]}
			drag.MulSelf(drag.Length())
			drag.MulSelf(-0.5 * constants.AirDensity * crossSectionArea * constants.SphereDragCoefficient)
			f.AddSelf(&drag)
			if s.Mob.OnGround {
				// Kinetic friction
				drag.From(v)
				g := concepts.Vector3{0, 0, constants.Gravity * s.Mob.Mass}
				drag.MulSelf(-s.FloorFriction * s.FloorNormal.Dot(&g))
				f.AddSelf(&drag)
			}
			//log.Printf("%v\n", drag)
		}
	}

	if !s.FloorMaterial.Nil() && s.Mob.Pos.Now[2] <= s.BottomZ.Now {
		s.ControllerSet.Act(s.TargetEntity, s.FloorMaterial, "Contact")
	}
	if !s.CeilMaterial.Nil() && s.Mob.Pos.Now[2] >= s.TopZ.Now {
		s.ControllerSet.Act(s.TargetEntity, s.CeilMaterial, "Contact")
	}

	mobTop := s.Mob.Pos.Now[2] + s.Mob.Height
	floorZ, ceilZ := s.SlopedZNow(s.Mob.Pos.Now.To2D())

	s.Mob.OnGround = false
	if !s.FloorTarget.Nil() && mobTop < floorZ {
		s.ControllerSet.Act(s.SourceEntity, s.TargetEntity, "Exit")
		s.ControllerSet.Act(s.SourceEntity, s.FloorTarget, "Enter")
		s.Sector = core.SectorFromDb(s.FloorTarget)
		_, ceilZ = s.Sector.SlopedZNow(s.Mob.Pos.Now.To2D())
		s.Mob.Pos.Now[2] = ceilZ - s.Mob.Height - 1.0
	} else if !s.FloorTarget.Nil() && s.Mob.Pos.Now[2] <= floorZ && s.Mob.Vel.Now[2] > 0 {
		s.Mob.Vel.Now[2] = constants.PlayerJumpForce
	} else if s.FloorTarget.Nil() && s.Mob.Pos.Now[2] <= floorZ {
		dist := s.FloorNormal[2] * (floorZ - s.Mob.Pos.Now[2])
		delta := s.FloorNormal.Mul(dist)
		s.Mob.Vel.Now.AddSelf(delta)
		s.Mob.Pos.Now.AddSelf(delta)
		s.Mob.OnGround = true
	}

	if !s.CeilTarget.Nil() && mobTop > ceilZ {
		s.ControllerSet.Act(s.SourceEntity, s.TargetEntity, "Exit")
		s.ControllerSet.Act(s.SourceEntity, s.CeilTarget, "Enter")
		s.Sector = core.SectorFromDb(s.CeilTarget)
		floorZ, _ = s.Sector.SlopedZNow(s.Mob.Pos.Now.To2D())
		s.Mob.Pos.Now[2] = floorZ - s.Mob.Height + 1.0
	} else if s.CeilTarget.Nil() && mobTop >= ceilZ {
		dist := -s.CeilNormal[2] * (mobTop - ceilZ + 1.0)
		delta := s.CeilNormal.Mul(dist)
		s.Mob.Vel.Now.AddSelf(delta)
		s.Mob.Pos.Now.AddSelf(delta)
	}
}
