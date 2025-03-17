// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"encoding/json"
	"log"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

// TODO: Be more flexible with what we can delete/cut/copy/paste. Should be
// possible to only grab a "hi" part of a segment, for example. For now,
// just grab EntityRefs
type Copy struct {
	state.IEditor

	Selected      *selection.Selection
	Saved         map[ecs.Entity]any
	ClipboardData string
}

func (a *Copy) Activate() {
	a.Saved = make(map[ecs.Entity]any)
	a.Selected = selection.NewSelectionClone(a.State().SelectedObjects)

	for _, obj := range a.Selected.Exact {
		// Don't copy/paste active players
		if p := behaviors.GetPlayer(obj.Universe, obj.Entity); p != nil && !p.Spawn {
			continue
		}
		a.Saved[obj.Entity] = obj.Serialize()
	}

	bytes, err := json.MarshalIndent(a.Saved, "", "  ")

	if err != nil {
		panic(err)
	}

	a.ClipboardData = string(bytes)

	a.Redo()
	a.ActionFinished(false, false, false)
}

func (a *Copy) Undo() {}

func (a *Copy) Redo() {
	log.Printf("%v\n", a.ClipboardData)
	a.IEditor.SetContent(a.ClipboardData)
}
