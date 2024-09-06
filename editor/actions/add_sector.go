// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// Declare conformity with interfaces
var _ fyne.Draggable = (*AddSector)(nil)
var _ desktop.Hoverable = (*AddSector)(nil)
var _ desktop.Mouseable = (*AddSector)(nil)

type AddSector struct {
	AddEntity
	Sector *core.Sector
}

func (a *AddSector) Act() {
	a.SetMapCursor(desktop.CrosshairCursor)
	a.Mode = "AddSector"
	a.SelectObjects(true, selection.SelectableFromSector(a.Sector))
	//set cursor
}

func (a *AddSector) Dragged(d *fyne.DragEvent) {
	a.MouseMoved(&desktop.MouseEvent{PointEvent: d.PointEvent})
}

func (a *AddSector) DragEnd() {
	a.MouseUp(&desktop.MouseEvent{})
}

func (a *AddSector) MouseDown(evt *desktop.MouseEvent) {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	a.Mode = "AddSectorSegment"

	seg := core.SectorSegment{}
	seg.Construct(a.State().ECS, nil)
	seg.Sector = a.Sector
	seg.HiSurface.Material = controllers.DefaultMaterial(a.State().ECS)
	seg.LoSurface.Material = controllers.DefaultMaterial(a.State().ECS)
	seg.Surface.Material = controllers.DefaultMaterial(a.State().ECS)
	seg.P = *a.WorldGrid(&a.State().MouseDownWorld)

	segs := a.Sector.Segments
	if len(segs) > 0 {
		seg.Prev = segs[len(segs)-1]
		seg.Next = segs[0]
		seg.Next.Prev = &seg
		seg.Prev.Next = &seg
	}

	a.Sector.Segments = append(segs, &seg)
	a.Sector.Recalculate()
}

func (a *AddSector) MouseIn(evt *desktop.MouseEvent) {}
func (a *AddSector) MouseOut()                       {}

func (a *AddSector) MouseMoved(evt *desktop.MouseEvent) {
	if a.Mode != "AddSectorSegment" {
		return
	}

	segs := a.Sector.Segments
	seg := segs[len(segs)-1]
	seg.P = *a.WorldGrid(&a.State().MouseWorld)
}

func (a *AddSector) MouseUp(evt *desktop.MouseEvent) {
	if a.Mode != "AddSectorSegment" {
		return
	}
	a.Mode = "AddSector"

	segs := a.Sector.Segments
	if len(segs) > 1 {
		first := segs[0]
		last := segs[len(segs)-1]
		if last.P.Sub(&first.P).Length() < state.SegmentSelectionEpsilon {
			a.State().Lock.Lock()
			a.Sector.Segments = segs[:(len(segs) - 1)]
			a.State().Modified = true
			a.Sector.Recalculate()
			a.State().Lock.Unlock()
			a.ActionFinished(false, true, true)
		}
	}
	// TODO: right-mouse button end
}

func (a *AddSector) Status() string {
	return "Click to place " + a.Mode
}
