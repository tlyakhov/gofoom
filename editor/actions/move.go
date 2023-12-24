package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2/driver/desktop"
)

type Move struct {
	state.IEditor

	Selected []any
	Original []concepts.Vector3
	Delta    concepts.Vector2
}

func (a *Move) Iterate(arr []any,
	bodyFunc func(*core.Body),
	sectorFunc func(*core.Sector),
	segmentFunc func(*core.Segment)) {
	for _, obj := range arr {
		switch target := obj.(type) {
		case *concepts.EntityRef:
			if body := core.BodyFromDb(target); body != nil {
				bodyFunc(body)
			}
			if sector := core.SectorFromDb(target); sector != nil {
				sectorFunc(sector)
			}

		case *core.Segment:
			segmentFunc(target)
		}
	}
}

func (a *Move) OnMouseDown(evt *desktop.MouseEvent) {
	a.SetMapCursor(desktop.PointerCursor)

	a.Selected = make([]any, len(a.State().SelectedObjects))
	copy(a.Selected, a.State().SelectedObjects)
	a.Original = make([]concepts.Vector3, len(a.Selected))

	j := 0
	segmentFunc := func(seg *core.Segment) {
		a.Original = append(a.Original, concepts.Vector3{})
		seg.P.To3D(&a.Original[j])
		j++
	}
	a.Iterate(a.Selected, func(body *core.Body) {
		a.Original = append(a.Original, concepts.Vector3{})
		a.Original[j] = body.Pos.Original
		j++
	}, func(sector *core.Sector) {
		for _, segment := range sector.Segments {
			segmentFunc(segment)
		}
	}, segmentFunc)
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
	j := 0
	segmentFunc := func(seg *core.Segment) {
		seg.P = *a.WorldGrid(a.Original[j].To2D().Add(&a.Delta))
		a.State().DB.NewControllerSet().Act(seg.Sector.Ref(), concepts.ControllerRecalculate)
		j++
	}
	a.Iterate(a.Selected, func(body *core.Body) {
		body.Pos.Original = *a.WorldGrid3D(a.Original[j].Add(a.Delta.To3D(new(concepts.Vector3))))
		body.Pos.Reset()
		a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
		j++
	}, func(sector *core.Sector) {
		for _, segment := range sector.Segments {
			segmentFunc(segment)
		}
	}, segmentFunc)
}
func (a *Move) Cancel() {}
func (a *Move) Frame()  {}

func (a *Move) Undo() {
	j := 0
	segmentFunc := func(seg *core.Segment) {
		seg.P = *a.Original[j].To2D()
		a.State().DB.NewControllerSet().Act(seg.Sector.Ref(), concepts.ControllerRecalculate)
		j++
	}
	a.Iterate(a.Selected, func(body *core.Body) {
		body.Pos.Original = a.Original[j]
		body.Pos.Reset()
		a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
		j++
	}, func(sector *core.Sector) {
		for _, segment := range sector.Segments {
			segmentFunc(segment)
		}
	}, segmentFunc)
}
func (a *Move) Redo() {
	a.Act()
}

func (a *Move) RequiresLock() bool { return true }
