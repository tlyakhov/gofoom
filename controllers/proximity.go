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
	*behaviors.Proximity
}

func init() {
	ecs.Types().RegisterController(&ProximityController{}, 100)
}

func (pc *ProximityController) ComponentID() ecs.ComponentID {
	return behaviors.ProximityCID
}

func (pc *ProximityController) Methods() ecs.ControllerMethod {
	return ecs.ControllerAlways
}

func (pc *ProximityController) Target(target ecs.Attachable) bool {
	pc.Proximity = target.(*behaviors.Proximity)
	return pc.IsActive()
}

func (pc *ProximityController) proximity(entity ecs.Entity) {
	if sector := core.GetSector(pc.ECS, entity); sector != nil {
		for _, pvs := range sector.PVS {
			for _, body := range pvs.Bodies {
				if sector.Center.Dist2(&body.Pos.Now) < pc.Range*pc.Range {
					BodySectorScript(pc.Scripts, body, sector)
				}
			}
		}
		return
	}
	if body := core.GetBody(pc.ECS, entity); body != nil && body.SectorEntity != 0 {
		container := body.Sector()
		for _, sector := range container.PVS {
			if sector.Center.Dist2(&body.Pos.Now) < pc.Range*pc.Range {
				BodySectorScript(pc.Scripts, body, sector)
			}
		}
	}
}

func (pc *ProximityController) Always() {
	// Is the target itself a body or sector?
	pc.proximity(pc.Entity)

	for entity := range pc.Entities {
		pc.proximity(entity)
	}
}
