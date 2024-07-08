// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2/driver/desktop"
)

type Transform struct {
	state.IEditor

	Selected  *core.Selection
	MouseDown concepts.Vector2
	Delta     concepts.Vector2
	Angle     float64
	Mode      string
}

func (a *Transform) OnMouseDown(evt *desktop.MouseEvent) {
	a.SetMapCursor(desktop.PointerCursor)

	a.Selected = core.NewSelectionClone(a.State().SelectedObjects)

	a.State().Lock.Lock()
	a.Selected.SavePositions()
	a.State().Lock.Unlock()
}

func (a *Transform) OnMouseMove() {
	a.MouseDown.From(&a.State().MouseDownWorld)
	a.Delta = *a.State().MouseWorld.Sub(&a.MouseDown)
	a.Angle = a.Delta[0] / 60.0
	if a.State().KeysDown[desktop.KeyShiftLeft] && a.State().KeysDown[desktop.KeyAltLeft] {
		a.Mode = "rotate"
	} else if a.State().KeysDown[desktop.KeyShiftLeft] {
		a.Mode = "rotate-constrained"
	} else {
		a.Mode = "translate"
	}
	a.Act()
}

func (a *Transform) OnMouseUp() {
	a.State().Lock.Lock()
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
	a.State().Lock.Unlock()
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}
func (a *Transform) Act() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	a.Selected.LoadPositions()
	m := &concepts.Matrix2{}
	m.SetIdentity()
	switch a.Mode {
	case "rotate":
		m = m.RotateBasis(&a.MouseDown, a.Angle)
	case "rotate-constrained":
		factor := math.Pi * 0.25
		m = m.RotateBasis(&a.MouseDown, math.Round(a.Angle/factor)*factor)
	default:
		m.Translate(a.WorldGrid(&a.Delta))
	}
	for _, s := range a.Selected.Exact {
		s.Transform(m)
		s.Recalculate()
	}
}
func (a *Transform) Cancel() {}
func (a *Transform) Frame()  {}

func (a *Transform) Undo() {
	a.Selected.LoadPositions()
	for _, s := range a.Selected.Exact {
		s.Recalculate()
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
func (a *Transform) Redo() {
	a.Act()
}

func (a *Transform) RequiresLock() bool { return true }

func (a *Transform) Status() string {
	return ""
}
