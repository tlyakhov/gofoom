package actions

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gdk"
)

type Pan struct {
	state.IEditor

	Mode        string
	OriginalPos concepts.Vector2
	Delta       concepts.Vector2
}

func (a *Pan) OnMouseDown(button *gdk.EventButton) {
	a.Mode = "PanStart"
	a.OriginalPos = a.State().Pos
	a.SetMapCursor("all-scroll")
}

func (a *Pan) OnMouseMove() {
	a.Mode = "Panning"
	a.Delta = *a.State().Mouse.Sub(&a.State().MouseDown)
	a.State().Pos = *a.OriginalPos.Sub(a.Delta.Mul(1.0 / a.State().Scale))
}

func (a *Pan) OnMouseUp() {
	a.ActionFinished(false)
}
func (a *Pan) Act()    {}
func (a *Pan) Cancel() {}
func (a *Pan) Frame()  {}

func (a *Pan) Undo() {
	a.State().Pos = a.OriginalPos
}
func (a *Pan) Redo() {
	a.State().Pos = *a.OriginalPos.Sub(&a.Delta)
}
