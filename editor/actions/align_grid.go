package actions

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gdk"
)

type AlignGrid struct {
	state.IEditor
	PrevA, PrevB, A, B concepts.Vector2
}

func (a *AlignGrid) OnMouseDown(button *gdk.EventButton) {}
func (a *AlignGrid) OnMouseMove()                        {}
func (a *AlignGrid) Frame()                              {}
func (a *AlignGrid) Act()                                {}

func (a *AlignGrid) OnMouseUp() {
	a.PrevA, a.PrevB = a.State().MapView.GridA, a.State().MapView.GridB
	a.A = a.WorldGrid(a.State().MouseDownWorld)
	a.B = a.WorldGrid(a.State().MouseWorld)
	if a.A.Dist2(a.B) < 0.001 {
		a.A = concepts.Vector2{}
		a.B = concepts.Vector2{X: 0, Y: 1}
	}
	a.State().MapView.GridA, a.State().MapView.GridB = a.A, a.B
	a.ActionFinished(false)
}

func (a *AlignGrid) Cancel() {
	a.ActionFinished(true)
}

func (a *AlignGrid) Undo() {
	a.State().MapView.GridA, a.State().MapView.GridB = a.PrevA, a.PrevB
}
func (a *AlignGrid) Redo() {
	a.State().MapView.GridA, a.State().MapView.GridB = a.A, a.B
}
