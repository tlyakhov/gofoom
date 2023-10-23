package actions

import (
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gdk"
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

func (a *MoveSurface) Get(sector *core.Sector) *float64 {
	if a.Slope {
		if a.Floor {
			return &sector.FloorSlope
		} else {
			return &sector.CeilSlope
		}
	} else {
		if a.Floor {
			return &sector.BottomZ.Original
		} else {
			return &sector.TopZ.Original
		}
	}
}

func (a *MoveSurface) Act() {
	a.Original = make([]float64, len(a.State().SelectedObjects))
	for i, obj := range a.State().SelectedObjects {
		switch p := obj.(type) {
		case core.Sector:
			a.Original[i] = *a.Get(p)
			*a.Get(p) += a.Delta
		case *state.MapPoint:
			a.Original[i] = *a.Get(p.Sector)
			*a.Get(p.Sector) += a.Delta
		}

	}
	a.State().World.Recalculate()
	a.State().Modified = true
	a.ActionFinished(false)
}

func (a *MoveSurface) Undo() {
	for i, obj := range a.State().SelectedObjects {
		if sector, ok := obj.(core.Sector); ok {
			*a.Get(sector) = a.Original[i]
		} else if p, ok := obj.(state.MapPoint); ok {
			*a.Get(p.Sector) = a.Original[i]
		}
	}
	a.State().World.Recalculate()
}
func (a *MoveSurface) Redo() {
	for _, obj := range a.State().SelectedObjects {
		if sector, ok := obj.(core.Sector); ok {
			*a.Get(sector) += a.Delta
		} else if p, ok := obj.(state.MapPoint); ok {
			*a.Get(p.Sector) += a.Delta
		}
	}
	a.State().World.Recalculate()
}
