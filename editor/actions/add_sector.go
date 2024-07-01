// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/driver/desktop"
)

type AddSector struct {
	AddEntity
	Sector *core.Sector
}

func (a *AddSector) Act() {
	a.SetMapCursor(desktop.CrosshairCursor)
	a.Mode = "AddSector"
	a.SelectObjects(true, core.SelectableFromSector(a.Sector))
	//set cursor
}

func (a *AddSector) OnMouseDown(evt *desktop.MouseEvent) {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	a.Mode = "AddSectorSegment"

	seg := core.SectorSegment{}
	seg.Construct(nil)
	seg.Sector = a.Sector
	seg.HiSurface.Material = controllers.DefaultMaterial(a.State().DB)
	seg.LoSurface.Material = controllers.DefaultMaterial(a.State().DB)
	seg.Surface.Material = controllers.DefaultMaterial(a.State().DB)
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
func (a *AddSector) OnMouseMove() {
	if a.Mode != "AddSectorSegment" {
		return
	}

	segs := a.Sector.Segments
	seg := segs[len(segs)-1]
	seg.P = *a.WorldGrid(&a.State().MouseWorld)
}

func (a *AddSector) OnMouseUp() {
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
