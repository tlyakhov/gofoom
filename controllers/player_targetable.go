// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
)

type PlayerTargetableController struct {
	ecs.BaseController
	*behaviors.PlayerTargetable
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &PlayerTargetableController{} }, 75)
}

func (ptc *PlayerTargetableController) ComponentID() ecs.ComponentID {
	return behaviors.PlayerTargetableCID
}

func (ptc *PlayerTargetableController) Methods() ecs.ControllerMethod {
	return ecs.ControllerRecalculate |
		ecs.ControllerLoaded
}

func (ptc *PlayerTargetableController) EditorPausedMethods() ecs.ControllerMethod {
	return ecs.ControllerRecalculate |
		ecs.ControllerLoaded
}

func (ptc *PlayerTargetableController) Target(target ecs.Attachable) bool {
	ptc.PlayerTargetable = target.(*behaviors.PlayerTargetable)
	return ptc.PlayerTargetable.IsActive()
}

var playerTargetableScriptParams = []core.ScriptParam{
	{Name: "body", TypeName: "*core.Body"},
	{Name: "player", TypeName: "*behaviors.Player"},
}

func (ptc *PlayerTargetableController) Recalculate() {
	if !ptc.Frob.IsEmpty() {
		ptc.Frob.Params = playerTargetableScriptParams
		ptc.Frob.Compile()
	}
	if !ptc.Selected.IsEmpty() {
		ptc.Selected.Params = playerTargetableScriptParams
		ptc.Selected.Compile()
	}
	if !ptc.UnSelected.IsEmpty() {
		ptc.UnSelected.Params = playerTargetableScriptParams
		ptc.UnSelected.Compile()
	}
}
