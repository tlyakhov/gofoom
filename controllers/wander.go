package controllers

import (
	"math"
	"math/rand"
	"strconv"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
)

type WanderController struct {
	concepts.BaseController
	*behaviors.Wander
	Body *core.Body
}

func init() {
	concepts.DbTypes().RegisterController(&WanderController{})
}

func (wc *WanderController) Methods() concepts.ControllerMethod {
	return concepts.ControllerAlways
}

func (wc *WanderController) Target(target *concepts.EntityRef) bool {
	wc.TargetEntity = target
	wc.Wander = behaviors.WanderFromDb(target)
	wc.Body = core.BodyFromDb(target)
	return wc.Wander != nil && wc.Wander.Active && wc.Body != nil && wc.Body.Active
}

func (wc *WanderController) Always() {
	// Applying an impulse directly is helpful for objects without mass.
	f := concepts.Vector3{}
	f[1], f[0] = math.Sincos(wc.Body.Angle.Now * concepts.Deg2rad)
	f.MulSelf(wc.Force)
	wc.Body.Force.AddSelf(&f)

	if wc.NextSector.Nil() {
		wc.NextSector = wc.Body.SectorEntityRef
	}

	if wc.Timestamp-wc.LastTurn > int64(300+rand.Intn(100)) {
		name := "wc_" + strconv.FormatUint(wc.TargetEntity.Entity, 10)
		a := new(concepts.Animation[float64])
		a.SetDB(wc.DB)
		a.Construct(nil)
		a.Name = name
		a.Target = &wc.Body.Angle
		a.Start = wc.Body.Angle.Now
		// Bias towards the center of the sector
		start := wc.Body.Angle.Now + rand.Float64()*60 - 30
		end := start
		if sector := core.SectorFromDb(wc.NextSector); sector != nil {
			end = wc.Body.Angle2DTo(&sector.Center)
		}
		a.End = concepts.TweenAngles(start, end, 0.2, concepts.Lerp)

		a.Duration = 300
		a.TweeningFunc = concepts.EaseInOut
		a.Style = concepts.AnimationStyleOnce
		wc.AttachAnimation(name, a)
		wc.LastTurn = wc.Timestamp
	}
	if wc.Timestamp-wc.LastTarget > int64(5000+rand.Intn(5000)) {
		sector := wc.Body.Sector()
		var closestSegment *core.Segment
		closestDist := constants.MaxViewDistance
		for _, seg := range sector.Segments {
			if seg.AdjacentSector.Nil() || !seg.PortalIsPassable {
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
		wc.LastTarget = wc.Timestamp
	}
}
