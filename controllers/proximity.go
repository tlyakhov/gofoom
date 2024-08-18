// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
)

type ProximityController struct {
	ecs.BaseController
	*core.Body
}

func init() {
	ecs.Types().RegisterController(&ProximityController{}, 100)
}

func (pc *ProximityController) ComponentIndex() int {
	return core.BodyComponentIndex
}

func (pc *ProximityController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways
}

func (pc *ProximityController) Target(target ecs.Attachable) bool {
	pc.Body = target.(*core.Body)
	return pc.IsActive() && pc.SectorEntity != 0
}

func (pc *ProximityController) proximity(sector *core.Sector, body *core.Body) {
	// Consider the case where the sector entity has a proximity
	// component that includes the body as a valid scripting source
	if p := behaviors.ProximityFromDb(pc.DB, sector.Entity); p != nil && p.IsActive() {
		if sector.Center.Dist2(&body.Pos.Now) < p.Range*p.Range {
			BodySectorScript(p.Scripts, body, sector)
		}
	}

	// Consider the case where the body entity has a proximity
	// component that includes the sector as a valid scripting source
	if p := behaviors.ProximityFromDb(pc.DB, body.Entity); p != nil && p.IsActive() {
		if sector.Center.Dist2(&body.Pos.Now) < p.Range*p.Range {
			BodySectorScript(p.Scripts, body, sector)
		}
	}
}

func (pc *ProximityController) Always() {
	sector := pc.Sector()
	if sector == nil {
		return
	}
	for _, pvs := range sector.PVS {
		pc.proximity(pvs, pc.Body)
	}
}
