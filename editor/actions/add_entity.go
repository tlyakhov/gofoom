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

// Declare conformity with interfaces
var _ Placeable = (*AddEntity)(nil)

type AddEntity struct {
	Place
	ecs.Entity

	Components       ecs.ComponentTable
	ContainingSector *core.Sector
}

func (a *AddEntity) DetachFromSector() {
	if body := core.GetBody(a.Entity); body != nil {
		if body.SectorEntity == 0 {
			return
		}
		delete(body.Sector().Bodies, a.Entity)
	}
}

func (a *AddEntity) AttachToSector() {
	if body := core.GetBody(a.Entity); body != nil {
		if a.ContainingSector != nil {
			body.SectorEntity = a.ContainingSector.Entity
			a.ContainingSector.Bodies[a.Entity] = body
		}
	}
	if seg := core.GetInternalSegment(a.Entity); seg != nil {
		seg.AttachToSectors()
	}
}

func (a *AddEntity) BeginPoint(m fyne.KeyModifier, button desktop.MouseButton) bool {
	if !a.Place.BeginPoint(m, button) {
		return false
	}

	return a.Point()
}
func (a *AddEntity) Point() bool {
	a.Place.Point()

	// The rest of this is special sauce for components that need to be linked
	// up to specific sectors (depending on where they're placed)
	body := core.GetBody(a.Entity)
	seg := core.GetInternalSegment(a.Entity)
	if body == nil && seg == nil {
		return true
	}

	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	worldGrid := a.WorldGrid(&a.State().MouseWorld)

	arena := ecs.ArenaFor[core.Sector](core.SectorCID)
	for i := range arena.Cap() {
		sector := arena.Value(i)
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
		body.Pos.Spawn[0] = worldGrid[0]
		body.Pos.Spawn[1] = worldGrid[1]
		if a.ContainingSector != nil {
			floorZ, ceilZ := a.ContainingSector.ZAt(dynamic.Spawn, worldGrid)
			body.Pos.Spawn[2] = (floorZ + ceilZ) / 2
		}
		body.Pos.ResetToSpawn()
	}
	return true
}

func (a *AddEntity) EndPoint() bool {
	if !a.Place.EndPoint() {
		return false
	}
	a.State().Lock.Lock()
	ecs.ActAllControllers(ecs.ControllerRecalculate)
	a.State().Modified = true
	a.State().Lock.Unlock()
	a.ActionFinished(false, true, a.Components.Get(core.SectorCID) != nil)
	return true
}

func (a *AddEntity) Activate() {
	a.Entity = ecs.NewEntity()
	for i, component := range a.Components {
		if component == nil {
			continue
		}
		ecs.Attach(component.ComponentID(), a.Entity, &a.Components[i])
	}
	a.SelectObjects(true, selection.SelectableFromEntity(a.Entity))
}
func (a *AddEntity) Cancel() {
	a.State().Lock.Lock()
	a.DetachFromSector()
	if a.Entity != 0 {
		ecs.Delete(a.Entity)
		a.Entity = 0
	}
	a.State().Lock.Unlock()
	a.SelectObjects(true)
	a.ActionFinished(true, true, false)
}
func (a *AddEntity) Undo() {
	a.DetachFromSector()
	ecs.Delete(a.Entity)
}
func (a *AddEntity) Redo() {
	a.Activate()
	a.AttachToSector()
}

func (a *AddEntity) Status() string {
	return "Click to place entity"
}

func (a *AddEntity) Serialize() map[string]any {
	return nil
}
