package actions

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/editor/state"
)

type MoveSurface struct {
	state.IEditor

	Original []float64
	Floor    bool
	Slope    bool
	Delta    float64
}

func (a *MoveSurface) OnMouseDown(button *gdk.EventButton) {}
func (a *MoveSurface) OnMouseMove()                        {}
func (a *MoveSurface) OnMouseUp()                          {}
func (a *MoveSurface) Cancel()                             {}
func (a *MoveSurface) Frame()                              {}

func (a *MoveSurface) Get(sector *core.PhysicalSector) *float64 {
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

func (a *MoveSurface) Act() {
	a.Original = make([]float64, len(a.State().SelectedObjects))
	for i, obj := range a.State().SelectedObjects {
		if sector, ok := obj.(core.AbstractSector); ok {
			a.Original[i] = *a.Get(sector.Physical())
			*a.Get(sector.Physical()) += a.Delta
		} else if p, ok := obj.(state.MapPoint); ok {
			a.Original[i] = *a.Get(p.Sector.Physical())
			*a.Get(p.Sector.Physical()) += a.Delta
		}
	}
	a.State().World.Recalculate()
	a.State().Modified = true
	a.ActionFinished(false)
}

func (a *MoveSurface) Undo() {
	for i, obj := range a.State().SelectedObjects {
		if sector, ok := obj.(core.AbstractSector); ok {
			*a.Get(sector.Physical()) = a.Original[i]
		} else if p, ok := obj.(state.MapPoint); ok {
			*a.Get(p.Sector.Physical()) = a.Original[i]
		}
	}
	a.State().World.Recalculate()
}
func (a *MoveSurface) Redo() {
	for _, obj := range a.State().SelectedObjects {
		if sector, ok := obj.(core.AbstractSector); ok {
			*a.Get(sector.Physical()) += a.Delta
		} else if p, ok := obj.(state.MapPoint); ok {
			*a.Get(p.Sector.Physical()) += a.Delta
		}
	}
	a.State().World.Recalculate()
}
