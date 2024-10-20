// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// Declare conformity with interfaces
var _ Placeable = (*AddSector)(nil)

type AddSector struct {
	AddEntity
	Sector *core.Sector
}

func (a *AddSector) Act() {
	a.AddEntity.Act()
	a.SetMapCursor(desktop.CrosshairCursor)
}

func (a *AddSector) BeginPoint(m fyne.KeyModifier, button desktop.MouseButton) bool {
	if !a.AddEntity.BeginPoint(m, button) {
		return false
	}
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	switch a.Mode {
	case "Begin":
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
	return true
}

func (a *AddSector) Point() bool {
	a.AddEntity.Point()
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	worldGrid := a.WorldGrid(&a.State().MouseWorld)

	switch a.Mode {
	case "Begin":
		segs := a.Sector.Segments
		seg := segs[len(segs)-1]
		seg.P = *worldGrid
	}
	return true
}

func (a *AddSector) EndPoint() bool {
	switch a.Mode {
	case "Begin":
		segs := a.Sector.Segments
		if len(segs) > 1 {
			first := segs[0]
			last := segs[len(segs)-1]
			if last.P.Sub(&first.P).Length() < state.SegmentSelectionEpsilon {
				a.State().Lock.Lock()
				a.Sector.Segments = segs[:(len(segs) - 1)]
				a.State().Lock.Unlock()
				return a.AddEntity.EndPoint()
			}
		}
	}
	return true
}

func (a *AddSector) Status() string {
	return "Click to place " + a.Mode
}
