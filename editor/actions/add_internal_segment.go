// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/dynamic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

type AddInternalSegment struct {
	AddEntity
	*core.InternalSegment
}

func (a *AddInternalSegment) Act() {
	a.AddEntity.Act()
	a.SetMapCursor(desktop.CrosshairCursor)
}

func (a *AddInternalSegment) BeginPoint(m fyne.KeyModifier, button desktop.MouseButton) bool {
	if !a.AddEntity.BeginPoint(m, button) {
		return false
	}
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	switch a.Mode {
	case "Begin":
		a.Mode = "AddInternalSegmentA"
		a.Surface.Material = controllers.DefaultMaterial(a.State().ECS)
		a.AttachToSector()
	case "AddInternalSegmentA":
	}
	return true
}

func (a *AddInternalSegment) Point() bool {
	a.AddEntity.Point()
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	worldGrid := a.WorldGrid(&a.State().MouseWorld)

	switch a.Mode {
	case "Begin", "AddInternalSegmentA":
		a.A.From(worldGrid)
		fallthrough
	case "AddInternalSegmentB":
		a.B.From(worldGrid)
	}
	if a.ContainingSector != nil {
		a.Bottom, a.Top = a.ContainingSector.ZAt(dynamic.DynamicOriginal, worldGrid)
	}
	a.Recalculate()
	return true
}

func (a *AddInternalSegment) EndPoint() bool {
	if !a.AddEntity.EndPoint() {
		return false
	}

	switch a.Mode {
	case "AddInternalSegmentA":
		a.State().Lock.Lock()
		a.A.From(a.WorldGrid(&a.State().MouseWorld))
		a.State().Lock.Unlock()
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
	return true
}

func (a *AddInternalSegment) Status() string {
	return "Click to place " + a.Mode
}
