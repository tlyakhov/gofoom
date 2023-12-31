package actions

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/driver/desktop"
)

type AlignGrid struct {
	state.IEditor
	PrevA, PrevB, A, B concepts.Vector2
}

func (a *AlignGrid) OnMouseDown(evt *desktop.MouseEvent) {}
func (a *AlignGrid) OnMouseMove()                        {}
func (a *AlignGrid) Frame()                              {}
func (a *AlignGrid) Act()                                {}

func (a *AlignGrid) OnMouseUp() {
	a.PrevA, a.PrevB = a.State().MapView.GridA, a.State().MapView.GridB
	a.A = *a.WorldGrid(&a.State().MouseDownWorld)
	a.B = *a.WorldGrid(&a.State().MouseWorld)
	if a.A.Dist2(&a.B) < 0.001 {
		a.A = concepts.Vector2{}
		a.B = concepts.Vector2{0, 1}
	}
	a.State().MapView.GridA, a.State().MapView.GridB = a.A, a.B
	a.ActionFinished(false, false, false)
}

func (a *AlignGrid) Cancel() {
	a.ActionFinished(true, false, false)
}

func (a *AlignGrid) Undo() {
	a.State().MapView.GridA, a.State().MapView.GridB = a.PrevA, a.PrevB
}
func (a *AlignGrid) Redo() {
	a.State().MapView.GridA, a.State().MapView.GridB = a.A, a.B
}
