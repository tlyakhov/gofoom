// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

type DeleteComponent struct {
	state.IEditor

	Components      map[ecs.Entity]ecs.Attachable
	SectorForEntity map[ecs.Entity]*core.Sector
}

func (a *DeleteComponent) Act() {
	a.SectorForEntity = make(map[ecs.Entity]*core.Sector)
	a.Redo()
	a.ActionFinished(false, true, false)
}

func (a *DeleteComponent) Undo() {
	for entity, component := range a.Components {
		a.State().ECS.UpsertTyped(entity, component)
		switch target := component.(type) {
		case *core.Body:
			entity := component.Base().Entity
			if a.SectorForEntity[entity] != nil {
				a.SectorForEntity[entity].Bodies[entity] = target
			}
		}
	}
	a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
}
func (a *DeleteComponent) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	for _, component := range a.Components {
		// TODO: Save material references
		switch target := component.(type) {
		case *core.Body:
			entity := component.Base().Entity
			if target.SectorEntity != 0 {
				a.SectorForEntity[entity] = target.Sector()
				delete(a.SectorForEntity[entity].Bodies, entity)
			}
		}
		a.State().ECS.DeleteByType(component)
	}
	a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
}
