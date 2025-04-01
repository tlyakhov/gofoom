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
	OldComponents    map[ecs.Entity]ecs.ComponentTable
	AddComponents    ecs.ComponentTable
	RemoveComponents containers.Set[ecs.ComponentID]
}

func (a *UpdateLinks) Activate() {
	a.Redo()
	a.ActionFinished(false, true, false)
}

func (a *UpdateLinks) Undo() {
	panic("Unimplemented")
}

func (a *UpdateLinks) attach(entity ecs.Entity) {
	u := a.State().Universe

	all := u.AllComponents(entity)
	oldComponents := make(ecs.ComponentTable, len(all))
	copy(oldComponents, all)
	a.OldComponents[entity] = oldComponents

	for _, oldComponent := range oldComponents {
		if oldComponent == nil {
			continue
		}
		cid := oldComponent.ComponentID()
		addComponent := a.AddComponents.Get(cid)
		if a.RemoveComponents.Contains(cid) || (addComponent != nil && addComponent != oldComponent) {
			u.DetachComponent(cid, entity)
		}
	}

	for _, addComponent := range a.AddComponents {
		if addComponent == nil {
			continue
		}
		cid := addComponent.ComponentID()
		oldComponent := oldComponents.Get(cid)
		if oldComponent == nil || addComponent != oldComponent {
			u.Attach(cid, entity, &addComponent)
		}
	}
}
func (a *UpdateLinks) Redo() {
	a.OldComponents = make(map[ecs.Entity]ecs.ComponentTable)
	for _, e := range a.Entities {
		if e != 0 {
			a.attach(e)
		}
	}
	a.State().Universe.ActAllControllers(ecs.ControllerRecalculate)
}
