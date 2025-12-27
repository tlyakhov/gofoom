// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

type UpdateLinks struct {
	state.Action

	Entities         ecs.EntityTable
	AddComponents    ecs.ComponentTable
	RemoveComponents containers.Set[ecs.ComponentID]
}

func (a *UpdateLinks) Activate() {
	for _, e := range a.Entities {
		if e != 0 {
			a.attach(e)
		}
	}
	ecs.ActAllControllers(ecs.ControllerPrecompute)
	a.State().Modified = true
	a.ActionFinished(false, true, false)
}

func (a *UpdateLinks) attach(entity ecs.Entity) {
	all := ecs.AllComponents(entity)
	oldComponents := make(ecs.ComponentTable, len(all))
	copy(oldComponents, all)

	for _, oldComponent := range oldComponents {
		if oldComponent == nil {
			continue
		}
		cid := oldComponent.ComponentID()
		addComponent := a.AddComponents.Get(cid)
		if a.RemoveComponents.Contains(cid) || (addComponent != nil && addComponent != oldComponent) {
			ecs.DetachComponent(cid, entity)
		}
	}

	for _, addComponent := range a.AddComponents {
		if addComponent == nil {
			continue
		}
		cid := addComponent.ComponentID()
		oldComponent := oldComponents.Get(cid)
		if oldComponent == nil || addComponent != oldComponent {
			ecs.Attach(cid, entity, &addComponent)
		}
	}
}
