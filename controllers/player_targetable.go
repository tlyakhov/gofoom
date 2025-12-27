// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"text/template"
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
	return ecs.ControllerPrecompute
}

func (ptc *PlayerTargetableController) EditorPausedMethods() ecs.ControllerMethod {
	return ecs.ControllerPrecompute
}

func (ptc *PlayerTargetableController) Target(target ecs.Component, e ecs.Entity) bool {
	ptc.Entity = e
	ptc.PlayerTargetable = target.(*behaviors.PlayerTargetable)
	return ptc.PlayerTargetable.IsActive()
}

var playerTargetableScriptParams = []core.ScriptParam{
	{Name: "body", TypeName: "*core.Body"},
	{Name: "player", TypeName: "*character.Player"},
	{Name: "carrier", TypeName: "*inventory.Carrier"},
}

func (ptc *PlayerTargetableController) Precompute() {
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
	var err error
	ptc.MessageTemplate, err = template.New("message").Funcs(ecs.FuncMap).Parse(ptc.Message)

	if err != nil {
		ptc.MessageTemplate, _ = template.New("error").Parse("Error: " + err.Error())
	}
}
