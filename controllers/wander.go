// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math"
	"math/rand"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"
)

type WanderController struct {
	ecs.BaseController
	*behaviors.Wander
	Body *core.Body
}

func init() {
	ecs.Types().RegisterController(&WanderController{}, 100)
}

func (wc *WanderController) ComponentID() ecs.ComponentID {
	return behaviors.WanderCID
}

func (wc *WanderController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways
}

func (wc *WanderController) Target(target ecs.Attachable) bool {
	wc.Wander = target.(*behaviors.Wander)
	wc.Body = core.GetBody(wc.Wander.ECS, wc.Wander.Entity)
	return wc.Wander.IsActive() && wc.Body.IsActive()
}

func (wc *WanderController) Always() {
	// Applying an impulse directly is helpful for objects without mass.
	f := concepts.Vector3{}
	f[1], f[0] = math.Sincos(wc.Body.Angle.Now * concepts.Deg2rad)
	f.MulSelf(wc.Force)
	wc.Body.Force.AddSelf(&f)

	if wc.NextSector == 0 {
		wc.NextSector = wc.Body.SectorEntity
	}

	if wc.ECS.Timestamp-wc.LastTurn > int64(300+rand.Intn(100)) {
		a := wc.Body.Angle.NewAnimation()
		a.Coordinates = dynamic.AnimationCoordinatesAbsolute
		a.Start = wc.Body.Angle.Now
		// Bias towards the center of the sector
		start := wc.Body.Angle.Now + rand.Float64()*60 - 30
		end := start
		if sector := core.GetSector(wc.Body.ECS, wc.NextSector); sector != nil {
			end = wc.Body.Angle2DTo(&sector.Center)
		}
		a.End = dynamic.TweenAngles(start, end, 0.2, dynamic.Lerp)

		a.Duration = 300
		a.TweeningFunc = dynamic.EaseInOut2
		a.Lifetime = dynamic.AnimationLifetimeOnce
		wc.LastTurn = wc.ECS.Timestamp
	}
	if wc.ECS.Timestamp-wc.LastTarget > int64(5000+rand.Intn(5000)) {
		sector := wc.Body.Sector()
		if sector == nil {
			return
		}
		var closestSegment *core.SectorSegment
		closestDist := constants.MaxViewDistance
		for _, seg := range sector.Segments {
			if seg.AdjacentSector == 0 || !seg.PortalIsPassable {
				continue
			}
			dist := seg.DistanceToPoint2(wc.Body.Pos.Now.To2D())
			if dist > closestDist {
				continue
			}
			closestDist = dist
			closestSegment = seg
		}
		if closestSegment != nil {
			wc.NextSector = closestSegment.AdjacentSector
		}
		wc.LastTarget = wc.ECS.Timestamp
	}
}
