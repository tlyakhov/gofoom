// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"encoding/json"
	"log"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/driver/desktop"
)

// TODO: Be more flexible with what we can delete/cut/copy/paste. Should be
// possible to only grab a "hi" part of a segment, for example. For now,
// just grab EntityRefs
type Copy struct {
	state.IEditor

	Selected      *core.Selection
	Saved         map[concepts.Entity]any
	ClipboardData string
}

func (a *Copy) Act() {
	a.Saved = make(map[concepts.Entity]any)
	a.Selected = core.NewSelectionClone(a.State().SelectedObjects)

	for _, obj := range a.Selected.Exact {
		// Don't copy/paste players
		if behaviors.PlayerFromDb(obj.DB, obj.Entity) != nil {
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
func (a *Copy) Cancel()                             {}
func (a *Copy) Frame()                              {}
func (a *Copy) OnMouseDown(evt *desktop.MouseEvent) {}
func (a *Copy) OnMouseMove()                        {}
func (a *Copy) OnMouseUp()                          {}

func (a *Copy) Undo() {
}
func (a *Copy) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	log.Printf("%v\n", a.ClipboardData)
	a.IEditor.SetContent(a.ClipboardData)
}

func (a *Copy) Status() string {
	return ""
}
