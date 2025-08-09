// SetEntityright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

type SetEntity struct {
	state.Action

	Selected    *selection.Selection
	StartEntity ecs.Entity
}

func (a *SetEntity) Activate() {
	a.Selected = selection.NewSelectionClone(a.State().SelectedObjects)

	for _, obj := range a.Selected.Exact {
		// Don't setentity for active players
		if p := character.GetPlayer(obj.Entity); p != nil && !p.Spawn {
			continue
		}
	}

	a.Redo()
	a.ActionFinished(false, false, false)
}

func (a *SetEntity) Undo() {

}

func (a *SetEntity) Redo() {
	/*
	   for _, s := range a.Selected.Exact {
	   }
	*/
}
