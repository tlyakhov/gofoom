package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"github.com/gotk3/gotk3/gdk"
)

type Move struct {
	state.IEditor

	Selected []any
	Original []concepts.Vector3
	Delta    concepts.Vector2
}

func (a *Move) OnMouseDown(button *gdk.EventButton) {
	a.SetMapCursor("move")

	a.Selected = []any{}
	for _, obj := range a.State().SelectedObjects {
		if sector, ok := obj.(core.Sector); ok {
			for _, seg := range sector.Segments {
				a.Selected = append(a.Selected, seg)
			}
		} else {
			a.Selected = append(a.Selected, obj)
		}
	}

	a.Original = make([]concepts.Vector3, len(a.Selected))
	for i, obj := range a.Selected {
		switch target := obj.(type) {
		case *core.Segment:
			target.P.To3D(&a.Original[i])
		case core.Body:
			a.Original[i] = target.Pos.Original
		}
	}
}

func (a *Move) OnMouseMove() {
	a.Delta = *a.State().MouseWorld.Sub(&a.State().MouseDownWorld)
	a.Act()
}

func (a *Move) OnMouseUp() {
	a.SelectObjects(a.State().SelectedObjects) // Updates properties.
	a.State().Modified = true
	a.ActionFinished(false)
}
func (a *Move) Act() {
	for i, obj := range a.Selected {
		switch target := obj.(type) {
		case *core.Segment:
			target.P = *a.WorldGrid(a.Original[i].To2D().Add(&a.Delta))
			a.State().DB.NewControllerSet().Act(target.Sector.Ref(), nil, concepts.ControllerRecalculate)
		case core.Body:
			target.Pos.Original = *a.WorldGrid3D(a.Original[i].Add(a.Delta.To3D(new(concepts.Vector3))))
			target.Pos.Reset()
			a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
		}
	}
}
func (a *Move) Cancel() {}
func (a *Move) Frame()  {}

func (a *Move) Undo() {
	for i, obj := range a.Selected {
		switch target := obj.(type) {
		case *core.Segment:
			target.P = *a.Original[i].To2D()
			a.State().DB.NewControllerSet().Act(target.Sector.Ref(), nil, concepts.ControllerRecalculate)
		case core.Body:
			target.Pos.Original = a.Original[i]
			target.Pos.Reset()
			a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
		}
	}
}
func (a *Move) Redo() {
	a.Act()
}
