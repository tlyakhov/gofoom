package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/core"
)

type MoveSurfaceAction struct {
	*Editor
	Original []float64
	Floor    bool
	Slope    bool
	Delta    float64
}

func (a *MoveSurfaceAction) OnMouseDown(button *gdk.EventButton) {}
func (a *MoveSurfaceAction) OnMouseMove()                        {}
func (a *MoveSurfaceAction) OnMouseUp()                          {}
func (a *MoveSurfaceAction) Cancel()                             {}
func (a *MoveSurfaceAction) Frame()                              {}

func (a *MoveSurfaceAction) Get(sector *core.PhysicalSector) *float64 {
	if a.Slope {
		if a.Floor {
			return &sector.FloorSlope
		} else {
			return &sector.CeilSlope
		}
	} else {
		if a.Floor {
			return &sector.BottomZ
		} else {
			return &sector.TopZ
		}
	}
}

func (a *MoveSurfaceAction) Act() {
	a.Original = make([]float64, len(a.SelectedObjects))
	for i, obj := range a.SelectedObjects {
		if sector, ok := obj.(core.AbstractSector); ok {
			a.Original[i] = *a.Get(sector.Physical())
			*a.Get(sector.Physical()) += a.Delta
		} else if p, ok := obj.(MapPoint); ok {
			a.Original[i] = *a.Get(p.Sector.Physical())
			*a.Get(p.Sector.Physical()) += a.Delta
		}
	}
	a.GameMap.Recalculate()
	a.RefreshPropertyGrid()
	a.ActionFinished()
}

func (a *MoveSurfaceAction) Undo() {
	for i, obj := range a.SelectedObjects {
		if sector, ok := obj.(core.AbstractSector); ok {
			*a.Get(sector.Physical()) = a.Original[i]
		} else if p, ok := obj.(MapPoint); ok {
			*a.Get(p.Sector.Physical()) = a.Original[i]
		}
	}
	a.GameMap.Recalculate()
	a.RefreshPropertyGrid()
}
func (a *MoveSurfaceAction) Redo() {
	for _, obj := range a.SelectedObjects {
		if sector, ok := obj.(core.AbstractSector); ok {
			*a.Get(sector.Physical()) += a.Delta
		} else if p, ok := obj.(MapPoint); ok {
			*a.Get(p.Sector.Physical()) += a.Delta
		}
	}
	a.GameMap.Recalculate()
	a.RefreshPropertyGrid()
}
