// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/dynamic"

	"fyne.io/fyne/v2/driver/desktop"
)

// Declare conformity with interfaces
var _ Placeable = (*AddInternalSegment)(nil)

type AddInternalSegment struct {
	AddEntity
	*core.InternalSegment
}

func (a *AddInternalSegment) Activate() {
	a.AddEntity.Activate()
	a.InternalSegment = a.Components.Get(core.InternalSegmentCID).(*core.InternalSegment)
	a.Surface.Material = controllers.DefaultMaterial(a.State().ECS)
	a.SetMapCursor(desktop.CrosshairCursor)
}

func (a *AddInternalSegment) Point() bool {
	a.AddEntity.Point()
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	worldGrid := a.WorldGrid(&a.State().MouseWorld)

	switch a.Mode {
	case "", "Placing":
		a.A.From(worldGrid)
		fallthrough
	case "AddInternalSegmentB":
		a.B.From(worldGrid)
	}
	if a.ContainingSector != nil {
		a.Bottom, a.Top = a.ContainingSector.ZAt(dynamic.DynamicSpawn, worldGrid)
	}
	a.Recalculate()
	return true
}

func (a *AddInternalSegment) EndPoint() bool {
	switch a.Mode {
	case "Placing":
		a.State().Lock.Lock()
		a.A.From(a.WorldGrid(&a.State().MouseWorld))
		a.State().Lock.Unlock()
		a.Mode = "AddInternalSegmentB"
	case "AddInternalSegmentB":
		a.B.From(a.WorldGrid(&a.State().MouseWorld))
		return a.AddEntity.EndPoint()
	}
	return true
}

func (a *AddInternalSegment) Status() string {
	return "Click to place " + a.Mode
}
