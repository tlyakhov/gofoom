// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"log"
	"tlyakhov/gofoom/components/character"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/editor/state"

	"gopkg.in/yaml.v3"
)

// TODO: Be more flexible with what we can delete/cut/copy/paste. Should be
// possible to only grab a "hi" part of a segment, for example. For now,
// just grab EntityRefs
type Copy struct {
	state.Action
	CutDelete *Delete

	Cut           bool
	Selected      *selection.Selection
	Saved         map[string]any
	ClipboardData string
}

func (a *Copy) Activate() {
	a.Saved = make(map[string]any)
	a.Selected = selection.NewSelectionClone(a.State().SelectedObjects)

	for _, obj := range a.Selected.Exact {
		// Don't copy/paste active players
		if p := character.GetPlayer(obj.Universe, obj.Entity); p != nil && !p.Spawn {
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

	if a.Cut {
		a.IEditor.SetContent(a.ClipboardData)
		a.CutDelete = &Delete{Action: state.Action{IEditor: a.IEditor}}
		// This will run ActionFinished
		a.CutDelete.Activate()
	} else {
		a.Redo()
		a.ActionFinished(false, false, false)
	}
}

func (a *Copy) Undo() {
	if a.Cut {
		a.CutDelete.Undo()
	}
}

func (a *Copy) Redo() {
	a.IEditor.SetContent(a.ClipboardData)
	if a.Cut {
		a.CutDelete.Redo()
	}
}
