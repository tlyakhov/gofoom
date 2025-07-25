// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

type AddComponent struct {
	state.Action

	Entities []ecs.Entity
	ID       ecs.ComponentID
}

func (a *AddComponent) Activate() {
	a.Redo()
	a.ActionFinished(false, true, a.ID == core.SectorCID)
}

func (a *AddComponent) Undo() {
	for _, entity := range a.Entities {
		ecs.DetachComponent(a.ID, entity)
	}
	ecs.ActAllControllers(ecs.ControllerRecalculate)
}
func (a *AddComponent) Redo() {
	for _, entity := range a.Entities {
		ecs.NewAttachedComponent(entity, a.ID)
	}
	ecs.ActAllControllers(ecs.ControllerRecalculate)
}

func (a *AddComponent) Construct(data map[string]any) {
}

func (a *AddComponent) Serialize() map[string]any {
	return nil
}
