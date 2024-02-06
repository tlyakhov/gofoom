package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/driver/desktop"
)

type RotateSegments struct {
	state.IEditor
}

func (a *RotateSegments) OnMouseDown(evt *desktop.MouseEvent) {}
func (a *RotateSegments) OnMouseMove()                        {}
func (a *RotateSegments) OnMouseUp()                          {}
func (a *RotateSegments) Cancel()                             {}
func (a *RotateSegments) Frame()                              {}

func (a *RotateSegments) Rotate(sector *core.Sector, backward bool) {
	length := len(sector.Segments)
	if length <= 1 {
		return
	}

	if backward {
		sector.Segments = append(sector.Segments[1:], sector.Segments[0])
	} else {
		sector.Segments = append([]*core.SectorSegment{sector.Segments[length-1]}, sector.Segments[:(length-1)]...)
	}
}
func (a *RotateSegments) Act() {
	a.Redo()
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}

func (a *RotateSegments) Undo() {
	for _, obj := range a.State().SelectedObjects {
		switch target := obj.(type) {
		case *concepts.EntityRef:
			if sector := core.SectorFromDb(target); sector != nil {
				a.Rotate(sector, true)
			}
		case *core.SectorSegment:
			a.Rotate(target.Sector, true)
		}
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
func (a *RotateSegments) Redo() {
	for _, obj := range a.State().SelectedObjects {
		switch target := obj.(type) {
		case *concepts.EntityRef:
			if sector := core.SectorFromDb(target); sector != nil {
				a.Rotate(sector, false)
			}
		case *core.SectorSegment:
			a.Rotate(target.Sector, false)
		}
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}

func (a *RotateSegments) RequiresLock() bool { return true }
