// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"log"
	"math"
	"math/rand/v2"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
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

	targetEntity   ecs.Entity
	target         *concepts.Vector3
	targetDelta    *concepts.Vector2
	distFromTarget float64
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &PursuerController{} }, 100)
}

func (pc *PursuerController) ComponentID() ecs.ComponentID {
	return behaviors.PursuerCID
}

func (pc *PursuerController) Methods() ecs.ControllerMethod {
	return ecs.ControllerFrame | ecs.ControllerPrecompute
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
	return true
}

func (pc *PursuerController) Precompute() {
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

func (pc *PursuerController) generateWeights() {
	if pc.target != nil {
		pc.targetDelta = pc.target.To2D().Sub(pc.Body.Pos.Now.To2D())
		pc.distFromTarget = pc.targetDelta.Length()
		pc.targetDelta[0] /= pc.distFromTarget
		pc.targetDelta[1] /= pc.distFromTarget
	}

	obstacles := pc.getObstacleBodies()
	for i, c := range pc.Candidates {
		// TODO: parameterize? obstacle avoidance distance
		angle := float64(i) * 360 / float64(len(pc.Candidates))
		c.Start.From(&pc.Body.Pos.Now)
		c.FromAngleAndLimit(angle, 0, constants.PursuitWallAvoidDistance)
		// First, weigh the target, if we have one
		if pc.target != nil {
			// Don't bias against opposite directions
			dot := c.Delta.To2D().Dot(pc.targetDelta)
			normal := c.Delta[0]*pc.targetDelta[1] - c.Delta[1]*pc.targetDelta[0]
			targetWeight := max(dot, 0)
			if pc.TargetInView {
				// This encourages enemies to strafe/retreat when they get close
				strafeWeight := 1 - math.Abs(-dot-0.3)
				// We need to bias the strafing to avoid equal opposing weights cancelling each other out
				if (pc.ClockwisePreference && normal < 0) || (!pc.ClockwisePreference && normal >= 0) {
					strafeWeight += 0.2
				} else {
					strafeWeight -= 0.2
				}
				c.Count++

				// Consistent hash to separate NPCs circling the player
				r := concepts.RngXorShift64(uint64(pc.Entity))
				strafeDistance := pc.StrafeDistance + float64(r%uint64(pc.StrafeDistance))
				strafeFactor := 1 - max(min(pc.distFromTarget/strafeDistance, 1), 0)
				c.Weight += strafeWeight*strafeFactor + targetWeight*(1-strafeFactor)
			} else {
				c.Weight += targetWeight
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

func (pc *PursuerController) targetOutOfView() {
	if len(pc.Breadcrumbs) == 0 {
		pc.target = nil
		return
	}

	b := pc.Breadcrumbs.PeekMax()
	pc.target = &b.Data.Pos
	radius := pc.Body.Size.Now[0] * 0.75
	if pc.Body.Pos.Now.DistSq(&b.Data.Pos) < radius*radius {
		r := pc.Breadcrumbs.RemoveMin()
		r.Data.Timestamp = ecs.Simulation.SimTimestamp
		r.Key = ecs.Simulation.SimTimestamp
		pc.Breadcrumbs.Insert(r)
	}
	if pc.TargetInView {
		log.Printf("Pursuer %v lost sight of %v!", pc.Entity, pc.targetEntity)
		pc.TargetInView = false
	}
}

func (pc *PursuerController) Frame() {
	if pc.ClockwiseSwitchTime == 0 || ecs.Simulation.SimTimestamp > pc.ClockwiseSwitchTime {
		pc.ClockwisePreference = !pc.ClockwisePreference
		pc.ClockwiseSwitchTime = ecs.Simulation.SimTimestamp + concepts.MillisToNanos(5000+10000*rand.Float64())
	}

	for i := range pc.Candidates {
		pc.Candidates[i] = &behaviors.Candidate{}
	}
	for len(pc.Breadcrumbs) > 0 {
		earliest := pc.Breadcrumbs.PeekMin()
		// 30 sec memory?
		// TODO: Parameterize this
		if ecs.Simulation.SimTimestamp-earliest.Key > concepts.MillisToNanos(30000) {
			pc.Breadcrumbs.RemoveMin()
		} else {
			break
		}
	}

	// For now, assume the pursued target is always the player
	player := pc.getPlayer()
	if player == nil {
		return
	}
	targetBody := core.GetBody(player.Entity)
	if targetBody == nil {
		return
	}
	pc.targetEntity = player.Entity
	pc.target = &targetBody.Pos.Now

	// defers are LIFO
	defer pc.pursue()
	defer pc.generateWeights()

	// For now, only use horizontal FOV to determine visibility, ignore vertical.
	pursuerToTargetAngle := pc.Body.Angle2DTo(pc.target)
	angleInFOV := concepts.NormalizeAngle(pursuerToTargetAngle - pc.Body.Angle.Now)
	if angleInFOV >= 180.0 {
		angleInFOV -= 360.0
	}
	if angleInFOV < -pc.FOV*0.5 || angleInFOV > pc.FOV*0.5 {
		pc.targetOutOfView()
		return
	}
	// We're within FOV, check for obstacles:
	ray := &concepts.Ray{Start: pc.Body.Pos.Now, End: *pc.target}
	ray.AnglesFromStartEnd()
	s, _ := Cast(ray, pc.Body.Sector(), pc.Entity, false)
	if s != nil && s.Entity != pc.targetEntity {
		// We're blocked by something
		//log.Printf("%v: %v", pc.Entity, s.String())
		pc.targetOutOfView()
		return
	}
	if !pc.TargetInView {
		log.Printf("Pursuer %v saw %v!", pc.Entity, pc.targetEntity)
	}
	pc.TargetInView = true
	// Only leave breadcrumbs every so often
	// TODO: parameterize this
	if len(pc.Breadcrumbs) > 0 {
		latest := pc.Breadcrumbs.PeekMax()
		if ecs.Simulation.SimTimestamp-latest.Key < concepts.MillisToNanos(500) {
			return
		}
		if targetBody.Pos.Now.EqualEpsilon(&latest.Data.Pos) {
			return
		}
	}
	// Leave breadcrumb
	pc.Breadcrumbs.Insert(gheap.HeapElement[int64, behaviors.Breadcrumb]{
		Key: ecs.Simulation.SimTimestamp,
		Data: behaviors.Breadcrumb{
			Pos:       targetBody.Pos.Now,
			Timestamp: ecs.Simulation.Timestamp,
		},
	})

	if len(pc.Breadcrumbs) > 10 {
		pc.Breadcrumbs.RemoveMin()
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
		return
	}
	delta.NormSelf()
	NpcMove(pc.Body, pc.ChaseSpeed, &delta, !pc.AlwaysFaceTarget)
	if pc.AlwaysFaceTarget && pc.target != nil {
		angle := math.Atan2(pc.targetDelta[1], pc.targetDelta[0]) * concepts.Rad2deg
		if pc.Body.Angle.Procedural {
			pc.Body.Angle.Input = angle
		} else {
			pc.Body.Angle.Now = angle
		}
	}
}
