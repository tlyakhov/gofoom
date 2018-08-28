package actions

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/editor/state"
)

type RotateSegments struct {
	state.IEditor
}

func (a *RotateSegments) OnMouseDown(button *gdk.EventButton) {}
func (a *RotateSegments) OnMouseMove()                        {}
func (a *RotateSegments) OnMouseUp()                          {}
func (a *RotateSegments) Cancel()                             {}
func (a *RotateSegments) Frame()                              {}

func (a *RotateSegments) Rotate(sector *core.PhysicalSector, backward bool) {
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
		if sector, ok := obj.(core.AbstractSector); ok {
			a.Rotate(sector.Physical(), true)
		} else if p, ok := obj.(state.MapPoint); ok {
			a.Rotate(p.Sector.Physical(), true)
		}
	}
	a.State().World.Recalculate()
}
func (a *RotateSegments) Redo() {
	for _, obj := range a.State().SelectedObjects {
		if sector, ok := obj.(core.AbstractSector); ok {
			a.Rotate(sector.Physical(), false)
		} else if p, ok := obj.(state.MapPoint); ok {
			a.Rotate(p.Sector.Physical(), false)
		}
	}
	a.State().World.Recalculate()
}
