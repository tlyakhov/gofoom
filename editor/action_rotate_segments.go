package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/core"
)

type RotateSegmentsAction struct {
	*Editor
}

func (a *RotateSegmentsAction) OnMouseDown(button *gdk.EventButton) {}
func (a *RotateSegmentsAction) OnMouseMove()                        {}
func (a *RotateSegmentsAction) OnMouseUp()                          {}
func (a *RotateSegmentsAction) Cancel()                             {}
func (a *RotateSegmentsAction) Frame()                              {}

func (a *RotateSegmentsAction) Rotate(sector *core.PhysicalSector, backward bool) {
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
func (a *RotateSegmentsAction) Act() {
	a.Redo()
	a.ActionFinished(false)
}

func (a *RotateSegmentsAction) Undo() {
	for _, obj := range a.SelectedObjects {
		if sector, ok := obj.(core.AbstractSector); ok {
			a.Rotate(sector.Physical(), true)
		} else if p, ok := obj.(MapPoint); ok {
			a.Rotate(p.Sector.Physical(), true)
		}
	}
	a.GameMap.Recalculate()
}
func (a *RotateSegmentsAction) Redo() {
	for _, obj := range a.SelectedObjects {
		if sector, ok := obj.(core.AbstractSector); ok {
			a.Rotate(sector.Physical(), false)
		} else if p, ok := obj.(MapPoint); ok {
			a.Rotate(p.Sector.Physical(), false)
		}
	}
	a.GameMap.Recalculate()
}
