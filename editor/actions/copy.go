// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"encoding/json"
	"fmt"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/driver/desktop"
)

// TODO: Be more flexible with what we can delete/cut/copy/paste. Should be
// possible to only grab a "hi" part of a segment, for example. For now,
// just grab EntityRefs
type Copy struct {
	state.IEditor

	Selected      []*core.Selectable
	Saved         map[uint64]any
	ClipboardData string
}

func (a *Copy) Act() {
	a.Saved = make(map[uint64]any)
	a.Selected = make([]*core.Selectable, len(a.State().SelectedObjects))
	copy(a.Selected, a.State().SelectedObjects)

	for _, obj := range a.Selected {
		a.Saved[obj.Ref.Entity] = obj.Serialize()
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

	fmt.Printf("%v\n", a.ClipboardData)
	a.IEditor.SetContent(a.ClipboardData)
}
