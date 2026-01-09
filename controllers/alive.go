// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
)

type AliveController struct {
	ecs.BaseController
	*behaviors.Alive
}

func init() {
	ecs.Types().RegisterController(func() ecs.Controller { return &AliveController{} }, 100)
}

func (a *AliveController) ComponentID() ecs.ComponentID {
	return behaviors.AliveCID
}

func (a *AliveController) Methods() ecs.ControllerMethod {
	return ecs.ControllerFrame | ecs.ControllerPrecompute
}

func (a *AliveController) Target(target ecs.Component, e ecs.Entity) bool {
	a.Entity = e
	a.Alive = target.(*behaviors.Alive)
	return a.IsActive()
}

var aliveScriptParams = []core.ScriptParam{
	{Name: "alive", TypeName: "*behaviors.Alive"},
	{Name: "onEntity", TypeName: "ecs.Entity"},
}

func (a *AliveController) Precompute() {
	if !a.Die.IsEmpty() {
		a.Die.Params = aliveScriptParams
		a.Die.Compile()
	}
	if !a.Live.IsEmpty() {
		a.Live.Params = aliveScriptParams
		a.Live.Compile()
	}
}

func (a *AliveController) Frame() {
	// TODO: Refactor cooldowns to be time-based rather than frames
	for source, d := range a.Damages {
		if d.Cooldown.Now <= 0 {
			d.Cooldown.Detach(ecs.Simulation)
			delete(a.Damages, source)
			continue
		}
		if d.Amount > 0 {
			if a.Health.Now > 0 {
				a.Health.Now -= d.Amount
			}
			d.Amount = 0
		}
		if d.Cooldown.IsProcedural() {
			d.Cooldown.Input--
		} else {
			d.Cooldown.Now--
		}
	}

	// TODO: We should also check our state when loading.
	if a.Health.Now <= 0 && a.Health.Prev > 0 {
		// We've become dead
		toDeactivate := [...]ecs.Component{
			ecs.GetComponent(a.Entity, core.MobileCID),
			ecs.GetComponent(a.Entity, behaviors.ActorCID),
			ecs.GetComponent(a.Entity, behaviors.WanderCID),
			ecs.GetComponent(a.Entity, behaviors.DoorCID),
		}
		for _, a := range toDeactivate {
			if a == nil {
				continue
			}
			a.Base().Flags &= ^ecs.ComponentActive
		}
		if a.Die.IsCompiled() {
			a.Die.Vars["onEntity"] = a.Entity
			a.Die.Vars["alive"] = a.Alive
			a.Die.Act()
		}
	} else if a.Health.Now > 0 && a.Health.Prev <= 0 {
		// We've come back to life!
		toReactivate := [...]ecs.Component{
			ecs.GetComponent(a.Entity, core.MobileCID),
			ecs.GetComponent(a.Entity, behaviors.ActorCID),
			ecs.GetComponent(a.Entity, behaviors.WanderCID),
			ecs.GetComponent(a.Entity, behaviors.DoorCID),
		}
		for _, a := range toReactivate {
			if a == nil {
				continue
			}
			a.Base().Flags |= ecs.ComponentActive
		}
		if a.Live.IsCompiled() {
			a.Live.Vars["onEntity"] = a.Entity
			a.Live.Vars["alive"] = a.Alive
			a.Live.Act()
		}
	}
}
