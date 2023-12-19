package controllers

import (
	"math"
	"math/rand"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
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
	f := wc.Wander.Dir
	f.MulSelf(wc.Force)
	wc.Body.Force.AddSelf(&f)

	if wc.Timestamp-wc.LastChange > int64(2000+rand.Intn(1000)) {
		wc.Dir[0] = rand.Float64() - 0.5
		wc.Dir[1] = rand.Float64() - 0.5
		wc.LastChange = wc.Timestamp
		wc.Dir.NormSelf()
		wc.Body.Angle.Now = math.Atan2(wc.Dir[1], wc.Dir[0]) * concepts.Rad2deg
	}
}
