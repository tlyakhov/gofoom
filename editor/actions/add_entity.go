// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

type AddEntity struct {
	Place
	ecs.Entity

	Components       []ecs.Attachable
	ContainingSector *core.Sector
}

func (a *AddEntity) DetachFromSector() {
	if body := core.GetBody(a.State().ECS, a.Entity); body != nil {
		if body.SectorEntity == 0 {
			return
		}
		delete(body.Sector().Bodies, a.Entity)
	}
	//a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
}

func (a *AddEntity) AttachAll() {
	for _, component := range a.Components {
		a.State().ECS.Attach(ecs.Types().ID(component), a.Entity, component)
	}
}

func (a *AddEntity) AttachToSector() {
	if body := core.GetBody(a.State().ECS, a.Entity); body != nil {
		if a.ContainingSector != nil {
			body.SectorEntity = a.ContainingSector.Entity
			a.ContainingSector.Bodies[a.Entity] = body
		}
	}
	if seg := core.GetInternalSegment(a.State().ECS, a.Entity); seg != nil {
		seg.AttachToSectors()
	}
	//a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
}

func (a *AddEntity) BeginPoint(m fyne.KeyModifier, button desktop.MouseButton) bool {
	if !a.Place.BeginPoint(m, button) {
		return false
	}

	return a.Point()
}
func (a *AddEntity) Point() bool {
	a.Place.Point()
	body := core.GetBody(a.State().ECS, a.Entity)
	seg := core.GetInternalSegment(a.State().ECS, a.Entity)
	if body == nil && seg == nil {
		return true
	}

	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	worldGrid := a.WorldGrid(&a.State().MouseWorld)

	col := ecs.ColumnFor[core.Sector](a.State().ECS, core.SectorCID)
	for i := range col.Cap() {
		sector := col.Value(i)
		if sector == nil {
			continue
		}
		if sector.IsPointInside2D(worldGrid) {
			a.ContainingSector = sector
			break
		}
	}

	a.DetachFromSector()
	a.AttachToSector()
	if body != nil {
		body.Pos.Original[0] = worldGrid[0]
		body.Pos.Original[1] = worldGrid[1]
		if a.ContainingSector != nil {
			floorZ, ceilZ := a.ContainingSector.ZAt(dynamic.DynamicOriginal, worldGrid)
			body.Pos.Original[2] = (floorZ + ceilZ) / 2
		}
		body.Pos.ResetToOriginal()
	}
	return true
}

func (a *AddEntity) EndPoint() bool {
	if !a.Place.EndPoint() {
		return false
	}
	a.State().Lock.Lock()
	a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
	a.State().Modified = true
	a.State().Lock.Unlock()
	a.ActionFinished(false, true, false)
	return true
}

func (a *AddEntity) Act() {
	a.SelectObjects(true, selection.SelectableFromEntity(a.State().ECS, a.Entity))
}
func (a *AddEntity) Cancel() {
	a.State().Lock.Lock()
	a.DetachFromSector()
	if a.Entity != 0 {
		a.State().ECS.DetachAll(a.Entity)
	}
	a.State().Lock.Unlock()
	a.SelectObjects(true)
	a.ActionFinished(true, true, false)
}
func (a *AddEntity) Undo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	a.DetachFromSector()
	a.State().ECS.DetachAll(a.Entity)
}
func (a *AddEntity) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	a.AttachAll()
	a.AttachToSector()
}

func (a *AddEntity) Status() string {
	return "Click to place entity"
}
