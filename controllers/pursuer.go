// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"math/rand/v2"
	"tlyakhov/gofoom/components/audio"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"github.com/srwiley/gheap"
)

// The approach here is inspired by
//
//	Game Endeavor - "The Trick I Used to Make Combat Fun! | Devlog"
//	https://www.youtube.com/watch?v=6BrZryMz-ac

type PursuerController struct {
	ecs.BaseController
	*behaviors.Pursuer
	Body   *core.Body
	Mobile *core.Mobile
	Alive  *behaviors.Alive

	faction string
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &PursuerController{} }, 100)
}

func (pc *PursuerController) ComponentID() ecs.ComponentID {
	return behaviors.PursuerCID
}

func (pc *PursuerController) Methods() ecs.ControllerMethod {
	return ecs.ControllerFrame
}

func (pc *PursuerController) Target(target ecs.Component, e ecs.Entity) bool {
	pc.Entity = e
	pc.Pursuer = target.(*behaviors.Pursuer)
	if !pc.Pursuer.IsActive() {
		return false
	}

	pc.Body = core.GetBody(pc.Entity)
	if pc.Body == nil || !pc.Body.IsActive() {
		return false
	}

	pc.Mobile = core.GetMobile(pc.Entity)
	if pc.Mobile == nil || !pc.Mobile.IsActive() {
		return false
	}
	pc.Alive = behaviors.GetAlive(pc.Entity)
	if pc.Alive != nil && pc.Alive.IsActive() {
		pc.faction = pc.Alive.Faction
	} else {
		pc.faction = ""
	}

	return true
}

func (pc *PursuerController) getPlayer() *character.Player {
	arena := ecs.ArenaFor[character.Player](character.PlayerCID)
	for i := range arena.Cap() {
		player := arena.Value(i)
		if player == nil || !player.IsActive() || behaviors.GetSpawner(player.Entity) != nil {
			continue
		}
		return player
	}
	return nil
}

func (pc *PursuerController) getObstacleBodies() []*core.Body {
	obstacles := make([]*core.Body, 0, 32)
	// Get potential obstacle bodies
	core.QuadTree.Root.RangeCircle(pc.Body.Pos.Now.To2D(), constants.PursuitNpcAvoidDistance, func(body *core.Body) bool {
		if !body.IsActive() || body == pc.Body {
			return true
		}
		if s := behaviors.GetSpawner(body.Entity); s != nil {
			// If this is a spawn point, skip it
			return true
		}

		mobile := core.GetMobile(body.Entity)
		if mobile != nil && mobile.IsActive() && mobile.CrBody == core.CollideNone {
			return true
		}
		// Ignore invisible bodies without a mobile
		if mobile == nil {
			if vis := materials.GetVisible(body.Entity); vis == nil {
				return true
			}
		}

		obstacles = append(obstacles, body)

		return true
	})
	return obstacles
}

func (pc *PursuerController) targetOutOfView(enemy *behaviors.PursuerEnemy) {
	if enemy.InView {
		//log.Printf("Pursuer %v lost sight of %v!", pc.Entity, enemy.Entity)
		pc.playSound(pc.SoundsTargetLost)
		enemy.InView = false
	}
	if len(enemy.Breadcrumbs) == 0 {
		enemy.Pos = nil
		return
	}

	b := enemy.Breadcrumbs.PeekMax()
	enemy.Pos = &b.Data.Pos
	radius := pc.Body.Size.Now[0] * 0.75
	if pc.Body.Pos.Now.DistSq(&b.Data.Pos) < radius*radius {
		r := enemy.Breadcrumbs.RemoveMin()
		r.Data.TargetTime = ecs.Simulation.SimTimestamp
		r.Key = ecs.Simulation.SimTimestamp
		enemy.Breadcrumbs.Insert(r)
	}
}

func (pc *PursuerController) targetEnemy(enemy *behaviors.PursuerEnemy) {
	// For now, only use horizontal FOV to determine visibility, ignore vertical.
	pursuerToTargetAngle := pc.Body.Angle2DTo(enemy.Pos)
	angleInFOV := concepts.NormalizeAngle(pursuerToTargetAngle - pc.Body.Angle.Now)
	if angleInFOV >= 180.0 {
		angleInFOV -= 360.0
	}
	if angleInFOV < -pc.FOV*0.5 || angleInFOV > pc.FOV*0.5 {
		pc.targetOutOfView(enemy)
		return
	}
	// We're within FOV, check for obstacles:
	ray := &concepts.Ray{Start: pc.Body.Pos.Now, End: *enemy.Pos}
	ray.AnglesFromStartEnd()
	s, _ := Cast(ray, pc.Body.Sector(), pc.Entity, false)
	if s != nil && s.Entity != enemy.Entity {
		// We're blocked by something
		//log.Printf("%v: %v", pc.Entity, s.String())
		pc.targetOutOfView(enemy)
		return
	}
	if !enemy.InView {
		//log.Printf("Pursuer %v saw %v!", pc.Entity, enemy.Entity)
		pc.playSound(pc.SoundsTargetSeen)
	}
	enemy.InView = true
	// Only leave breadcrumbs every so often
	// TODO: parameterize this
	if len(enemy.Breadcrumbs) > 0 {
		latest := enemy.Breadcrumbs.PeekMax()
		if ecs.Simulation.SimTimestamp-latest.Key < constants.PursuitBreadcrumbRateNs {
			return
		}
		radius := pc.Body.Size.Now[0] * 0.5
		if enemy.Body.Pos.Now.DistSq(&latest.Data.Pos) < radius*radius {
			return
		}
	}
	// Leave breadcrumb
	enemy.Breadcrumbs.Insert(gheap.HeapElement[int64, behaviors.Breadcrumb]{
		Key: ecs.Simulation.SimTimestamp,
		Data: behaviors.Breadcrumb{
			Pos:         enemy.Body.Pos.Now,
			TargetTime:  ecs.Simulation.Timestamp,
			CreatedTime: ecs.Simulation.SimTimestamp,
		},
	})

	// Remove breadcrumbs if we have too many
	if len(enemy.Breadcrumbs) > constants.PursuitMaxBreadcrumbs {
		enemy.Breadcrumbs.RemoveMin()
	}
}

func (pc *PursuerController) populateEnemies() {
	// Mark all existing enemies unvisited
	for _, enemy := range pc.Enemies {
		enemy.Visited = false
	}
	// Get potential enemies, update positions
	core.QuadTree.Root.RangeCircle(pc.Body.Pos.Now.To2D(), constants.PursuitEnemyTargetDistance, func(body *core.Body) bool {
		if !body.IsActive() || body == pc.Body {
			return true
		}
		var a *behaviors.Alive
		if a = behaviors.GetAlive(body.Entity); a == nil || !a.IsActive() {
			return true
		}
		if a.Faction == pc.faction || a.Health.Now <= 0 {
			// Skip entities of the same faction, or dead
			return true
		}
		if s := behaviors.GetSpawner(body.Entity); s != nil {
			// If this is a spawn point, skip it
			return true
		}

		var enemy *behaviors.PursuerEnemy
		var ok bool
		if enemy, ok = pc.Enemies[body.Entity]; !ok {
			enemy = &behaviors.PursuerEnemy{
				Entity: body.Entity,
			}
			pc.Enemies[enemy.Entity] = enemy
		}
		enemy.Pos = &body.Pos.Now
		enemy.Body = body
		enemy.Visited = true
		// Identify the position to target, either the enemy itself if in sight,
		// or a breadcrumb
		pc.targetEnemy(enemy)
		if enemy.Pos == nil {
			// Out of sight and no breadcrumbs!
			return true
		}
		enemy.Delta = &concepts.Vector2{
			enemy.Pos[0] - pc.Body.Pos.Now[0],
			enemy.Pos[1] - pc.Body.Pos.Now[1]}
		enemy.Dist = enemy.Delta.Length()
		enemy.Delta[0] /= enemy.Dist
		enemy.Delta[1] /= enemy.Dist

		return true
	})
	for e, enemy := range pc.Enemies {
		if !enemy.Visited {
			delete(pc.Enemies, e)
		}
	}

}

func (pc *PursuerController) targetWeight(c *behaviors.Candidate, enemy *behaviors.PursuerEnemy) float64 {
	// Don't bias against opposite directions
	dot := c.Delta.To2D().Dot(enemy.Delta)
	normal := c.Delta[0]*enemy.Delta[1] - c.Delta[1]*enemy.Delta[0]
	targetWeight := max(dot, 0)
	if !enemy.InView {
		return targetWeight
	}

	// This encourages pursuers to strafe/retreat when they get close to enemies
	strafeWeight := 1 - math.Abs(-dot-0.65)
	// We need to bias the strafing to avoid equal opposing weights cancelling each other out
	if (pc.ClockwisePreference && normal < 0) || (!pc.ClockwisePreference && normal >= 0) {
		strafeWeight += 0.2
	} else {
		strafeWeight -= 0.2
	}
	c.Count++

	// Consistent hash to separate NPCs circling a single enemy
	r := concepts.RngXorShift64(uint64(pc.Entity))
	strafeDistance := pc.StrafeDistance + float64(r%uint64(pc.StrafeDistance))
	strafeFactor := 1 - max(min(enemy.Dist/strafeDistance, 1), 0)
	return strafeWeight*strafeFactor + targetWeight*(1-strafeFactor)
}

func (pc *PursuerController) generateWeights() {
	obstacles := pc.getObstacleBodies()
	for i, c := range pc.Candidates {
		// TODO: parameterize? obstacle avoidance distance
		angle := float64(i) * 360 / float64(len(pc.Candidates))
		c.Start.From(&pc.Body.Pos.Now)
		c.FromAngleAndLimit(angle, 0, constants.PursuitWallAvoidDistance)
		c.Weight = 0
		c.Count = 0

		for _, enemy := range pc.Enemies {
			if enemy.Pos != nil {
				c.Weight += pc.targetWeight(c, enemy)
			}
		}

		// Other NPCs
		for _, body := range obstacles {
			delta := pc.Body.Pos.Now.To2D().Sub(body.Pos.Now.To2D())
			dist := delta.Length()
			delta[0] /= dist
			delta[1] /= dist
			weight := c.Delta.To2D().Dot(delta)
			distWeight := 1 - max(min(dist/constants.PursuitNpcAvoidDistance, 1), 0)
			weight = 1 - math.Abs(weight-0.65)
			c.Count++
			c.Weight += weight * distWeight
		}

		// Next, identify non-body obstacles
		s, hit := Cast(&c.Ray, pc.Body.Sector(), pc.Entity, true)

		// Geometry (e.g. walls)
		if s != nil {
			c.Count++
			c.Weight -= 1.0 - min(hit.Dist(&pc.Body.Pos.Now)/constants.PursuitWallAvoidDistance, 1)
		}
	}
}

func (pc *PursuerController) pruneBreadcrumbs() {
	for _, enemy := range pc.Enemies {
		for len(enemy.Breadcrumbs) > 0 {
			earliest := enemy.Breadcrumbs.PeekMin()
			// 45 sec memory?
			// TODO: Parameterize this
			if ecs.Simulation.SimTimestamp-earliest.Data.CreatedTime > concepts.MillisToNanos(45000) {
				enemy.Breadcrumbs.RemoveMin()
			} else {
				break
			}
		}
	}
}

func (pc *PursuerController) Frame() {
	if pc.ClockwiseSwitchTime == 0 || ecs.Simulation.SimTimestamp > pc.ClockwiseSwitchTime {
		pc.ClockwisePreference = !pc.ClockwisePreference
		pc.ClockwiseSwitchTime = ecs.Simulation.SimTimestamp + concepts.MillisToNanos(5000+10000*rand.Float64())
	}
	pc.pruneBreadcrumbs()
	pc.populateEnemies()
	pc.generateWeights()
	pc.pursue()
	pc.fire()
}

func (pc *PursuerController) fire() {
	if ecs.Simulation.SimTimestamp < pc.NextFireTime {
		return
	}
	for _, enemy := range pc.Enemies {
		if !enemy.InView {
			continue
		}
		// Fire
		if carrier := inventory.GetCarrier(pc.Entity); carrier != nil {
			if carrier.SelectedWeapon != 0 {
				if weapon := inventory.GetWeapon(carrier.SelectedWeapon); weapon != nil {
					weapon.Intent = inventory.WeaponFire
					pc.NextFireTime = ecs.Simulation.SimTimestamp + concepts.MillisToNanos(pc.FireDelay*0.5) + concepts.MillisToNanos(rand.Float64()*pc.FireDelay)
				}
			}
		}

		break
	}
}

func (pc *PursuerController) faceTarget() {
	newAngle := pc.Body.Angle.Now
	closest := math.Inf(1)
	closestAngle := newAngle
	// Pick closest angle to our current angle
	for _, enemy := range pc.Enemies {
		if enemy.Pos == nil {
			continue
		}
		a := math.Atan2(enemy.Delta[1], enemy.Delta[0]) * concepts.Rad2deg
		concepts.MinimizeAngleDistance(pc.Body.Angle.Now, &a)
		d := math.Abs(a - pc.Body.Angle.Now)
		if d < closest {
			closest = d
			closestAngle = a
		}
	}
	newAngle = dynamic.Lerp(newAngle, closestAngle, 0.25)
	if pc.Body.Angle.Procedural {
		pc.Body.Angle.Input = newAngle
	} else {
		pc.Body.Angle.Now = newAngle
	}
}

func (pc *PursuerController) idleFrame() {
	//log.Printf("Nowhere to go! Delta: %v", delta.StringHuman(2))
	if actor := behaviors.GetActor(pc.Entity); actor != nil {
		actor.Flags |= ecs.ComponentActive
	}
	if ecs.Simulation.SimTimestamp > pc.NextIdleBark {
		if pc.NextIdleBark != 0 {
			pc.playSound(pc.SoundsIdle)
		}
		pc.NextIdleBark = ecs.Simulation.SimTimestamp + concepts.MillisToNanos(10000+rand.Float64()*10000)

	}
}

func (pc *PursuerController) pursue() {
	delta := concepts.Vector3{}
	allNegative := true
	for _, c := range pc.Candidates {
		delta[0] += c.Delta[0] * c.Weight
		delta[1] += c.Delta[1] * c.Weight
		if c.Weight > 0 {
			allNegative = false
		}
	}
	if delta.Zero() || len(pc.Candidates) == 0 || allNegative {
		pc.idleFrame()
		return
	}
	if actor := behaviors.GetActor(pc.Entity); actor != nil {
		actor.Flags &= ^ecs.ComponentActive
	}
	delta.NormSelf()
	NpcMove(pc.Body, pc.ChaseSpeed, &delta, !pc.AlwaysFaceTarget)
	if pc.AlwaysFaceTarget {
		pc.faceTarget()
	}
}

func (pc *PursuerController) playSound(sounds ecs.EntityTable) {
	sound := ecs.Entity(0)
	r := rand.IntN(sounds.Len())
	for _, e := range sounds {
		if e == 0 {
			continue
		}
		if r <= 0 {
			sound = e
			break
		}
		r--
	}
	if sound == 0 {
		return
	}

	event, _ := audio.PlaySound(sound, pc.Body.Entity, pc.Body.Entity.String()+" voice", true)
	if event != nil {
		// TODO: Parameterize this
		event.Offset[2] = pc.Body.Size.Now[1] * 0.3
		event.Offset[1] = 0
		event.Offset[0] = 0
	}
}
