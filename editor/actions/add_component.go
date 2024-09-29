// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

type AddComponent struct {
	state.IEditor

	Entities []ecs.Entity
	ID       ecs.ComponentID
}

func (a *AddComponent) Act() {
	a.Redo()
	a.ActionFinished(false, true, a.ID == core.SectorCID)
}

func (a *AddComponent) Undo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	for _, entity := range a.Entities {
		a.State().ECS.DeleteComponent(a.ID, entity)
	}
	a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
}
func (a *AddComponent) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	for _, entity := range a.Entities {
		a.State().ECS.NewAttachedComponent(entity, a.ID)
	}
	a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
}
