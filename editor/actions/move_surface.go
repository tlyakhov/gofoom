// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

type SurfaceMode int

const (
	SurfaceModeLevel SurfaceMode = iota
	SurfaceModePhi
	SurfaceModeTheta
)

type MoveSurface struct {
	state.Action

	Original []any
	Mode     SurfaceMode
	Floor    bool
	Delta    float64
}

func (a *MoveSurface) Activate() {
	a.Original = make([]any, 0)
	for _, s := range a.State().SelectedObjects.Exact {
		if s.Sector == nil {
			continue
		}

		plane := &s.Sector.Top
		if a.Floor {
			plane = &s.Sector.Bottom
		}

		switch a.Mode {
		case SurfaceModePhi:
			a.Original = append(a.Original, plane.Normal)
			plane.Normal.SphericalQuantIncPhi(5*concepts.Deg2rad, a.Delta)
			plane.Normal.NormSelf()
		case SurfaceModeTheta:
			a.Original = append(a.Original, plane.Normal)
			plane.Normal.SphericalQuantIncTheta(5*concepts.Deg2rad, a.Delta)
			plane.Normal.NormSelf()
		default:
			a.Original = append(a.Original, plane.Z.Spawn)
			plane.Z.Spawn += a.Delta
			plane.Z.ResetToSpawn()
		}
	}
	ecs.ActAllControllers(ecs.ControllerRecalculate)
	a.State().Modified = true
	a.ActionFinished(false, true, false)
}
