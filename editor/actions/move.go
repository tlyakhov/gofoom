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

	Selected []*core.Selectable
	Original []concepts.Vector3
	Delta    concepts.Vector2
}

func (a *Move) OnMouseDown(evt *desktop.MouseEvent) {
	a.SetMapCursor(desktop.PointerCursor)

	a.Selected = make([]*core.Selectable, len(a.State().SelectedObjects))
	copy(a.Selected, a.State().SelectedObjects)
	a.Original = make([]concepts.Vector3, 0, len(a.Selected))

	a.State().Lock.Lock()
	for _, s := range a.Selected {
		a.Original = append(a.Original, s.SavePositions()...)
	}
	a.State().Lock.Unlock()
}

func (a *Move) OnMouseMove() {
	a.Delta = *a.State().MouseWorld.Sub(&a.State().MouseDownWorld)
	a.Act()
}

func (a *Move) OnMouseUp() {
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}
func (a *Move) Act() {
	a.State().Lock.Lock()

	m := &concepts.Matrix2{}
	m.SetIdentity()
	m.Translate(a.WorldGrid(&a.Delta))
	i := 0
	for _, s := range a.Selected {
		i += s.LoadPositions(a.Original[i:])
		s.Transform(m)
		s.Recalculate()
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
	a.State().Lock.Unlock()
}
func (a *Move) Cancel() {}
func (a *Move) Frame()  {}

func (a *Move) Undo() {
	i := 0
	for _, s := range a.Selected {
		i += s.LoadPositions(a.Original[i:])
		s.Recalculate()
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
func (a *Move) Redo() {
	a.Act()
}

func (a *Move) RequiresLock() bool { return true }
