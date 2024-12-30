// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// Declare conformity with interfaces
var _ fyne.Draggable = (*Pan)(nil)
var _ desktop.Hoverable = (*Pan)(nil)
var _ desktop.Mouseable = (*Pan)(nil)

type Pan struct {
	state.IEditor

	Mode        string
	OriginalPos concepts.Vector2
	Delta       concepts.Vector2
}

func (a *Pan) begin() {
	if a.Mode != "" {
		return
	}
	a.Mode = "Panning"
	a.OriginalPos = a.State().Pos
	a.SetMapCursor(desktop.VResizeCursor)
}

func (a *Pan) end() {
	if a.Mode != "Panning" {
		return
	}
	a.Mode = ""
	a.ActionFinished(false, false, false)
}

func (a *Pan) MouseDown(evt *desktop.MouseEvent) {
	a.begin()
}

func (a *Pan) MouseIn(evt *desktop.MouseEvent) {}
func (a *Pan) MouseOut()                       {}

func (a *Pan) MouseMoved(evt *desktop.MouseEvent) {
	a.begin()
	a.Delta = *a.State().Mouse.Sub(&a.State().MouseDown)
	a.State().Pos = *a.OriginalPos.Sub(a.Delta.Mul(1.0 / a.State().Scale))
}

func (a *Pan) Dragged(evt *fyne.DragEvent) {
	a.begin()
	a.Delta = *a.State().Mouse.Sub(&a.State().MouseDown)
	a.State().Pos = *a.OriginalPos.Sub(a.Delta.Mul(1.0 / a.State().Scale))
}

func (a *Pan) DragEnd() {
	a.end()
}

func (a *Pan) MouseUp(evt *desktop.MouseEvent) {
	a.end()
}
func (a *Pan) Activate() {}
func (a *Pan) Cancel()   {}

func (a *Pan) Undo() {
	a.State().Pos = a.OriginalPos
}
func (a *Pan) Redo() {
	a.State().Pos = *a.OriginalPos.Sub(&a.Delta)
}

func (a *Pan) RequiresLock() bool { return true }

func (a *Pan) Status() string {
	return "Panning"
}
