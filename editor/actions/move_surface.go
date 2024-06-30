// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/driver/desktop"
)

type MoveSurface struct {
	state.IEditor

	Original []float64
	Floor    bool
	Slope    bool
	Delta    float64
}

func (a *MoveSurface) OnMouseDown(evt *desktop.MouseEvent) {}
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
	a.State().Lock.Lock()

	a.Original = make([]float64, 0)
	for _, s := range a.State().SelectedObjects.Exact {
		if s.Sector == nil {
			continue
		}

		a.Original = append(a.Original, *a.Get(s.Sector))
		*a.Get(s.Sector) += a.Delta
		s.Sector.BottomZ.Reset()
		s.Sector.TopZ.Reset()
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
	a.State().Modified = true
	a.State().Lock.Unlock()
	a.ActionFinished(false, true, false)
}

func (a *MoveSurface) Undo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	i := 0
	for _, s := range a.State().SelectedObjects.Exact {
		if s.Sector == nil {
			continue
		}

		*a.Get(s.Sector) = a.Original[i]
		s.Sector.BottomZ.Reset()
		s.Sector.TopZ.Reset()
		i++
	}

	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
func (a *MoveSurface) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	for _, s := range a.State().SelectedObjects.Exact {
		if s.Sector == nil {
			continue
		}

		*a.Get(s.Sector) += a.Delta
		s.Sector.BottomZ.Reset()
		s.Sector.TopZ.Reset()
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
