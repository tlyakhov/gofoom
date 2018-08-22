package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/core"
)

type MoveSurfaceAction struct {
	*Editor
	Original []float64
	Floor    bool
	Delta    float64
}

func (a *MoveSurfaceAction) OnMouseDown(button *gdk.EventButton) {}
func (a *MoveSurfaceAction) OnMouseMove()                        {}
func (a *MoveSurfaceAction) OnMouseUp()                          {}
func (a *MoveSurfaceAction) Cancel()                             {}
func (a *MoveSurfaceAction) Frame()                              {}

func (a *MoveSurfaceAction) Act() {
	a.Original = make([]float64, len(a.SelectedObjects))
	for i, obj := range a.SelectedObjects {
		if sector, ok := obj.(core.AbstractSector); ok {
			if a.Floor {
				a.Original[i] = sector.Physical().BottomZ
				sector.Physical().BottomZ += a.Delta
			} else {
				a.Original[i] = sector.Physical().TopZ
				sector.Physical().TopZ += a.Delta
			}
		} else if p, ok := obj.(MapPoint); ok {
			if a.Floor {
				a.Original[i] = p.Sector.Physical().BottomZ
				p.Sector.Physical().BottomZ += a.Delta
			} else {
				a.Original[i] = p.Sector.Physical().TopZ
				p.Sector.Physical().TopZ += a.Delta
			}
		}
	}
	a.RefreshPropertyGrid()
	a.ActionFinished()
}

func (a *MoveSurfaceAction) Undo() {
	for i, obj := range a.SelectedObjects {
		if sector, ok := obj.(core.AbstractSector); ok {
			if a.Floor {
				sector.Physical().BottomZ = a.Original[i]
			} else {
				sector.Physical().TopZ = a.Original[i]
			}
		} else if p, ok := obj.(MapPoint); ok {
			if a.Floor {
				p.Sector.Physical().BottomZ = a.Original[i]
			} else {
				p.Sector.Physical().TopZ = a.Original[i]
			}
		}
	}
	a.RefreshPropertyGrid()
}
func (a *MoveSurfaceAction) Redo() {
	for _, obj := range a.SelectedObjects {
		if sector, ok := obj.(core.AbstractSector); ok {
			if a.Floor {
				sector.Physical().BottomZ += a.Delta
			} else {
				sector.Physical().TopZ += a.Delta
			}
		} else if p, ok := obj.(MapPoint); ok {
			if a.Floor {
				p.Sector.Physical().BottomZ += a.Delta
			} else {
				p.Sector.Physical().TopZ += a.Delta
			}
		}
	}
	a.RefreshPropertyGrid()
}
