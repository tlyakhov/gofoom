// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
)

type ScriptedController struct {
	ecs.BaseController
	*core.Scripted
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &ScriptedController{} }, 100)
}

func (sc *ScriptedController) ComponentID() ecs.ComponentID {
	return core.ScriptedCID
}

func (sc *ScriptedController) Methods() ecs.ControllerMethod {
	return ecs.ControllerFrame | ecs.ControllerRecalculate
}

func (sc *ScriptedController) Target(target ecs.Component, e ecs.Entity) bool {
	sc.Entity = e
	sc.Scripted = target.(*core.Scripted)
	return sc.Scripted.IsActive()
}

var scriptedScriptParams = []core.ScriptParam{
	{Name: "scripted", TypeName: "*core.Scripted"},
	{Name: "onEntity", TypeName: "ecs.Entity"},
	{Name: "args", TypeName: "[]string"},
}

func (sc *ScriptedController) Recalculate() {
	if !sc.OnFrame.IsEmpty() {
		sc.OnFrame.Params = scriptedScriptParams
		sc.OnFrame.Compile()
	}
}

func (sc *ScriptedController) Frame() {
	if !sc.OnFrame.IsCompiled() {
		return
	}
	sc.Apply(sc.Entity, func(e ecs.Entity) {
		sc.OnFrame.Vars["scripted"] = sc.Scripted
		sc.OnFrame.Vars["onEntity"] = e
		sc.OnFrame.Vars["args"] = sc.Args
		sc.OnFrame.Act()
	})
}
