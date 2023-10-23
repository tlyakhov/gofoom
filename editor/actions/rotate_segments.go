package actions

import (
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gdk"
)

type RotateSegments struct {
	state.IEditor
}

func (a *RotateSegments) OnMouseDown(button *gdk.EventButton) {}
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
		sector.Segments = append([]*core.Segment{sector.Segments[length-1]}, sector.Segments[:(length-1)]...)
	}
}
func (a *RotateSegments) Act() {
	a.Redo()
	a.State().Modified = true
	a.ActionFinished(false)
}

func (a *RotateSegments) Undo() {
	for _, obj := range a.State().SelectedObjects {
		if sector, ok := obj.(core.Sector); ok {
			a.Rotate(sector, true)
		} else if p, ok := obj.(state.MapPoint); ok {
			a.Rotate(p.Sector, true)
		}
	}
	a.State().World.Recalculate()
}
func (a *RotateSegments) Redo() {
	for _, obj := range a.State().SelectedObjects {
		if sector, ok := obj.(core.Sector); ok {
			a.Rotate(sector, false)
		} else if p, ok := obj.(state.MapPoint); ok {
			a.Rotate(p.Sector, false)
		}
	}
	a.State().World.Recalculate()
}
