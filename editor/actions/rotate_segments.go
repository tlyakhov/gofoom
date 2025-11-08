// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

type RotateSegments struct {
	state.Action
}

func (a *RotateSegments) Rotate(sector *core.Sector, backward bool) {
	length := len(sector.Segments)
	if length <= 1 {
		return
	}

	if backward {
		sector.Segments = append(sector.Segments[1:], sector.Segments[0])
	} else {
		sector.Segments = append([]*core.SectorSegment{sector.Segments[length-1]}, sector.Segments[:(length-1)]...)
	}
}
func (a *RotateSegments) Activate() {
	for _, s := range a.State().Selection.Exact {
		if s.Type == selection.SelectableBody || s.Type == selection.SelectableEntity {
			continue
		}
		a.Rotate(s.Sector, false)
	}
	ecs.ActAllControllers(ecs.ControllerRecalculate)
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}

func (a *RotateSegments) RequiresLock() bool { return true }

func (a *RotateSegments) Status() string {
	return ""
}
