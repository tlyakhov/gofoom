// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"log"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/editor/state"

	"gopkg.in/yaml.v3"
)

// TODO: Be more flexible with what we can delete/cut/copy/paste. Should be
// possible to only grab a "hi" part of a segment, for example. For now,
// just grab EntityRefs
type Copy struct {
	state.IEditor

	Selected      *selection.Selection
	Saved         map[string]any
	ClipboardData string
}

func (a *Copy) Activate() {
	a.Saved = make(map[string]any)
	a.Selected = selection.NewSelectionClone(a.State().SelectedObjects)

	for _, obj := range a.Selected.Exact {
		// Don't copy/paste active players
		if p := behaviors.GetPlayer(obj.Universe, obj.Entity); p != nil && !p.Spawn {
			continue
		}
		a.Saved[obj.Entity.Serialize(obj.Universe)] = obj.Serialize()
	}

	bytes, err := yaml.Marshal(a.Saved)

	if err != nil {
		log.Printf("Copy.Activate: error serializing to YAML: %v", err)
		a.ActionFinished(true, false, false)
		return
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
