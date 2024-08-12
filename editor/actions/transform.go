// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"log"
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// Declare conformity with interfaces
var _ fyne.Draggable = (*Transform)(nil)
var _ desktop.Hoverable = (*Transform)(nil)
var _ desktop.Mouseable = (*Transform)(nil)

type Transform struct {
	state.IEditor

	Selected  *core.Selection
	mouseDown concepts.Vector2
	Delta     concepts.Vector2
	Angle     float64
	Mode      string
}

func (a *Transform) begin() {
	if a.Mode != "" {
		return
	}

	a.SetMapCursor(desktop.PointerCursor)

	a.Selected = core.NewSelectionClone(a.State().SelectedObjects)
	a.Mode = "translate"

	a.State().Lock.Lock()
	a.Selected.SavePositions()
	a.State().Lock.Unlock()
}

func (a *Transform) end() {
	if a.Mode == "" {
		return
	}

	a.State().Lock.Lock()
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
	a.State().Lock.Unlock()
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}

func (a *Transform) moved(_ fyne.KeyModifier) {
	if a.Mode == "" {
		return
	}
	a.mouseDown.From(&a.State().MouseDownWorld)
	a.Delta = *a.State().MouseWorld.Sub(&a.mouseDown)
	log.Printf("Delta: %v, %v -> %v", a.Delta, a.mouseDown.StringHuman(), a.State().MouseWorld.StringHuman())
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

func (a *Transform) MouseMoved(evt *desktop.MouseEvent) {
	a.moved(evt.Modifier)
}

func (a *Transform) Dragged(evt *fyne.DragEvent) {
	a.begin()
	a.moved(fyne.KeyModifier(0))
}

func (a *Transform) DragEnd() {
	a.end()
}

func (a *Transform) MouseDown(evt *desktop.MouseEvent) {
	a.begin()
}

func (a *Transform) MouseUp(evt *desktop.MouseEvent) {
	a.end()
}

func (a *Transform) MouseIn(evt *desktop.MouseEvent) {
}

func (a *Transform) MouseOut() {
}

func (a *Transform) Act() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	a.Selected.LoadPositions()
	m := &concepts.Matrix2{}
	m.SetIdentity()
	switch a.Mode {
	case "rotate":
		m = m.RotateBasis(&a.mouseDown, a.Angle)
	case "rotate-constrained":
		factor := math.Pi * 0.25
		m = m.RotateBasis(&a.mouseDown, math.Round(a.Angle/factor)*factor)
	default:
		m.TranslateSelf(a.WorldGrid(&a.Delta))
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
