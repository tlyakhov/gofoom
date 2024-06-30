// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2/driver/desktop"
)

type Move struct {
	state.IEditor

	Selected *core.Selection
	Delta    concepts.Vector2
}

func (a *Move) OnMouseDown(evt *desktop.MouseEvent) {
	a.SetMapCursor(desktop.PointerCursor)

	a.Selected = core.NewSelectionClone(a.State().SelectedObjects)

	a.State().Lock.Lock()
	a.Selected.SavePositions()
	a.State().Lock.Unlock()
}

func (a *Move) OnMouseMove() {
	a.Delta = *a.State().MouseWorld.Sub(&a.State().MouseDownWorld)
	a.Act()
}

func (a *Move) OnMouseUp() {
	a.State().Lock.Lock()
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
	a.State().Lock.Unlock()
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}
func (a *Move) Act() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	a.Selected.LoadPositions()
	m := &concepts.Matrix2{}
	m.SetIdentity()
	m.Translate(a.WorldGrid(&a.Delta))
	for _, s := range a.Selected.Exact {
		s.Transform(m)
		s.Recalculate()
	}
}
func (a *Move) Cancel() {}
func (a *Move) Frame()  {}

func (a *Move) Undo() {
	a.Selected.LoadPositions()
	for _, s := range a.Selected.Exact {
		s.Recalculate()
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
func (a *Move) Redo() {
	a.Act()
}

func (a *Move) RequiresLock() bool { return true }
