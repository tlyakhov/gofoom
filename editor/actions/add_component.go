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
	for _, entity := range a.Entities {
		ecs.NewAttachedComponent(entity, a.ID)
	}
	ecs.ActAllControllers(ecs.ControllerRecalculate)
	a.ActionFinished(false, true, a.ID == core.SectorCID)
}

func (a *AddComponent) Construct(data map[string]any) {
}

func (a *AddComponent) Serialize() map[string]any {
	return nil
}
