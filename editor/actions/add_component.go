// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"
)

type AddComponent struct {
	state.IEditor

	Entities []concepts.Entity
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
		a.State().DB.Detach(a.Index, entity)
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
func (a *AddComponent) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	for _, entity := range a.Entities {
		a.State().DB.NewAttachedComponent(entity, a.Index)
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
