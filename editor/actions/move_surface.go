// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

type MoveSurface struct {
	state.IEditor

	Original []float64
	Floor    bool
	Slope    bool
	Delta    float64
}

func (a *MoveSurface) Get(sector *core.Sector) *float64 {
	if a.Slope {
		if a.Floor {
			//return &sector.FloorSlope
			return &sector.Bottom.Normal[2]
		} else {
			//return &sector.CeilSlope
			return &sector.Top.Normal[2]
		}
	} else {
		if a.Floor {
			return &sector.Bottom.Z.Spawn
		} else {
			return &sector.Top.Z.Spawn
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
		s.Sector.Bottom.Z.ResetToSpawn()
		s.Sector.Top.Z.ResetToSpawn()
	}
	a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
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
		s.Sector.Bottom.Z.ResetToSpawn()
		s.Sector.Top.Z.ResetToSpawn()
		i++
	}

	a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
}
func (a *MoveSurface) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	for _, s := range a.State().SelectedObjects.Exact {
		if s.Sector == nil {
			continue
		}

		*a.Get(s.Sector) += a.Delta
		s.Sector.Bottom.Z.ResetToSpawn()
		s.Sector.Top.Z.ResetToSpawn()
	}
	a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
}
