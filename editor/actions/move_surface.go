package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
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
		switch target := obj.(type) {
		case *concepts.EntityRef:
			if sector := core.SectorFromDb(target); sector != nil {
				a.Original[i] = *a.Get(sector)
				*a.Get(sector) += a.Delta
				sector.BottomZ.Reset()
				sector.TopZ.Reset()
			}
		case *core.Segment:
			a.Original[i] = *a.Get(target.Sector)
			*a.Get(target.Sector) += a.Delta
			target.Sector.BottomZ.Reset()
			target.Sector.TopZ.Reset()

		}

	}
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
	a.State().Modified = true
	a.ActionFinished(false, true, false)
}

func (a *MoveSurface) Undo() {
	for i, obj := range a.State().SelectedObjects {
		switch target := obj.(type) {
		case *concepts.EntityRef:
			if sector := core.SectorFromDb(target); sector != nil {
				*a.Get(sector) = a.Original[i]
				sector.BottomZ.Reset()
				sector.TopZ.Reset()
			}
		case *core.Segment:
			*a.Get(target.Sector) = a.Original[i]
			target.Sector.BottomZ.Reset()
			target.Sector.TopZ.Reset()
		}
	}
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
}
func (a *MoveSurface) Redo() {
	for _, obj := range a.State().SelectedObjects {
		switch target := obj.(type) {
		case *concepts.EntityRef:
			if sector := core.SectorFromDb(target); sector != nil {
				*a.Get(sector) += a.Delta
				sector.BottomZ.Reset()
				sector.TopZ.Reset()
			}
		case *core.Segment:
			*a.Get(target.Sector) += a.Delta
			target.Sector.BottomZ.Reset()
			target.Sector.TopZ.Reset()
		}
	}
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
}
