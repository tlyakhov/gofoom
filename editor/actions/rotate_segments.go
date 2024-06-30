// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/driver/desktop"
)

type RotateSegments struct {
	state.IEditor
}

func (a *RotateSegments) OnMouseDown(evt *desktop.MouseEvent) {}
func (a *RotateSegments) OnMouseMove()                        {}
func (a *RotateSegments) OnMouseUp()                          {}
func (a *RotateSegments) Cancel()                             {}
func (a *RotateSegments) Frame()                              {}

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
func (a *RotateSegments) Act() {
	a.Redo()
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}

func (a *RotateSegments) Undo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	for _, s := range a.State().SelectedObjects.Exact {
		if s.Type == core.SelectableBody || s.Type == core.SelectableEntityRef {
			continue
		}
		a.Rotate(s.Sector, true)
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
func (a *RotateSegments) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	for _, s := range a.State().SelectedObjects.Exact {
		if s.Type == core.SelectableBody || s.Type == core.SelectableEntityRef {
			continue
		}
		a.Rotate(s.Sector, false)
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}

func (a *RotateSegments) RequiresLock() bool { return true }
