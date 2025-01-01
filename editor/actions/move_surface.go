// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"math"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

type SurfaceMode int

const (
	SurfaceModeLevel SurfaceMode = iota
	SurfaceModeXYAngle
	SurfaceModeSlopeAngle
)

type MoveSurface struct {
	state.IEditor

	Original []any
	Mode     SurfaceMode
	Floor    bool
	Delta    float64
}

// Adapted from https://stackoverflow.com/questions/9038392/how-to-round-a-2d-vector-to-nearest-15-degrees?rq=4
func snapIncXY(v *concepts.Vector3, deg float64, inc float64) {
	l2d := math.Sqrt(v[0]*v[0] + v[1]*v[1])
	snapAngle := deg * concepts.Deg2rad
	angle := math.Atan2(v[1], v[0])
	newAngle := (math.Round(angle/snapAngle) + inc) * snapAngle
	v[0] = math.Cos(newAngle) * l2d
	v[1] = math.Sin(newAngle) * l2d
}

func snapIncSlope(v *concepts.Vector3, deg float64, inc float64) {
	l2d := math.Sqrt(v[0]*v[0] + v[1]*v[1])
	l3d := v.Length()
	snapAngle := deg * concepts.Deg2rad
	angle := math.Atan2(v[2], l2d)
	newAngle := (math.Round(angle/snapAngle) + inc) * snapAngle
	xyAngle := math.Atan2(v[1], v[0])
	v[0] = math.Cos(xyAngle) * math.Cos(newAngle) * l3d
	v[1] = math.Sin(xyAngle) * math.Cos(newAngle) * l3d
	v[2] = math.Sin(newAngle) * l3d
}

func (a *MoveSurface) Activate() {
	a.Redo()
	a.ActionFinished(false, true, false)
}

func (a *MoveSurface) Undo() {
	i := 0
	for _, s := range a.State().SelectedObjects.Exact {
		if s.Sector == nil {
			continue
		}

		plane := &s.Sector.Top
		if a.Floor {
			plane = &s.Sector.Bottom
		}

		switch a.Mode {
		case SurfaceModeXYAngle, SurfaceModeSlopeAngle:
			plane.Normal = a.Original[i].(concepts.Vector3)
		default:
			plane.Z.Spawn = a.Original[i].(float64)
			plane.Z.ResetToSpawn()
		}
		i++
	}

	a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
	a.State().Modified = true
}

func (a *MoveSurface) Redo() {
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
		case SurfaceModeXYAngle:
			a.Original = append(a.Original, plane.Normal)
			snapIncXY(&plane.Normal, 5, a.Delta)
			plane.Normal.NormSelf()
		case SurfaceModeSlopeAngle:
			a.Original = append(a.Original, plane.Normal)
			snapIncSlope(&plane.Normal, 5, a.Delta)
			plane.Normal.NormSelf()
		default:
			a.Original = append(a.Original, plane.Z.Spawn)
			plane.Z.Spawn += a.Delta
			plane.Z.ResetToSpawn()
		}
	}
	a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
	a.State().Modified = true
}
