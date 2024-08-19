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
	Index    int
}

func (a *AddComponent) Act() {
	a.Redo()
	a.ActionFinished(false, true, a.Index == core.SectorComponentIndex)
}

func (a *AddComponent) Undo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	for _, entity := range a.Entities {
		a.State().ECS.Detach(a.Index, entity)
	}
	a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
}
func (a *AddComponent) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	for _, entity := range a.Entities {
		a.State().ECS.NewAttachedComponent(entity, a.Index)
	}
	a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
}
