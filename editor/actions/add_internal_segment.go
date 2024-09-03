// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"fyne.io/fyne/v2/driver/desktop"
)

type AddInternalSegment struct {
	AddEntity
	*core.InternalSegment
}

func (a *AddInternalSegment) Act() {
	a.SetMapCursor(desktop.CrosshairCursor)
	a.Mode = "AddInternalSegment"
	a.SelectObjects(true, core.SelectableFromInternalSegment(a.InternalSegment))
	//set cursor
}

func (a *AddInternalSegment) OnMouseDown(evt *desktop.MouseEvent) {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	switch a.Mode {
	case "AddInternalSegment":
		a.Mode = "AddInternalSegmentA"
		a.Surface.Material = controllers.DefaultMaterial(a.State().ECS)
		a.AttachToSector()
	case "AddInternalSegmentA":
	}
}
func (a *AddInternalSegment) OnMouseMove() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	worldGrid := a.WorldGrid(&a.State().MouseWorld)

	col := ecs.ColumnFor[core.Sector](a.State().ECS, core.SectorCID)
	for i := range col.Length {
		sector := col.Value(i)
		if sector.IsPointInside2D(worldGrid) {
			a.ContainingSector = sector
			break
		}
	}

	a.DetachFromSector()
	a.AttachToSector()
	switch a.Mode {
	case "AddInternalSegment":
		fallthrough
	case "AddInternalSegmentA":
		a.A.From(worldGrid)
		fallthrough
	case "AddInternalSegmentB":
		a.B.From(worldGrid)
	}
	if a.ContainingSector != nil {
		a.Bottom, a.Top = a.ContainingSector.ZAt(dynamic.DynamicOriginal, worldGrid)
	}
	a.Recalculate()
}

func (a *AddInternalSegment) OnMouseUp() {
	switch a.Mode {
	case "AddInternalSegmentA":
		a.A.From(a.WorldGrid(&a.State().MouseWorld))
		a.Mode = "AddInternalSegmentB"
	case "AddInternalSegmentB":
		a.B.From(a.WorldGrid(&a.State().MouseWorld))
		a.Mode = "AddInternalSegment"
		a.State().Lock.Lock()
		a.State().Modified = true
		a.Recalculate()
		a.State().Lock.Unlock()
		a.ActionFinished(false, true, true)
	}
}

func (a *AddInternalSegment) Status() string {
	return "Click to place " + a.Mode
}
