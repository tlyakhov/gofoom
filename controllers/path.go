// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
)

type PathController struct {
	ecs.BaseController
	*core.Path
}

func init() {
	// Should run before everything
	ecs.Types().RegisterController(&PathController{}, 50)
}

func (sc *PathController) ComponentID() ecs.ComponentID {
	return core.PathCID
}

func (sc *PathController) Methods() ecs.ControllerMethod {
	return ecs.ControllerRecalculate
}

func (sc *PathController) Target(target ecs.Attachable) bool {
	sc.Path = target.(*core.Path)
	return sc.Path.IsActive()
}

func (sc *PathController) Recalculate() {
	sc.Path.Recalculate()
}
