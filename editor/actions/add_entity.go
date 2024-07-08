// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/driver/desktop"
)

type AddEntity struct {
	state.IEditor
	concepts.Entity

	Mode             string
	Components       []concepts.Attachable
	ContainingSector *core.Sector
}

func (a *AddEntity) DetachFromSector() {
	if body := core.BodyFromDb(a.State().DB, a.Entity); body != nil {
		if body.SectorEntity == 0 {
			return
		}
		delete(body.Sector().Bodies, a.Entity)
	}
	//a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}

func (a *AddEntity) AttachAll() {
	for index, component := range a.Components {
		a.State().DB.Attach(index, a.Entity, component)
	}
}

func (a *AddEntity) AttachToSector() {
	if body := core.BodyFromDb(a.State().DB, a.Entity); body != nil {
		if a.ContainingSector != nil {
			body.SectorEntity = a.ContainingSector.Entity
			a.ContainingSector.Bodies[a.Entity] = body
		}
	}
	if seg := core.InternalSegmentFromDb(a.State().DB, a.Entity); seg != nil {
		seg.AttachToSectors()
	}
	//a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}

func (a *AddEntity) OnMouseDown(evt *desktop.MouseEvent) {}

func (a *AddEntity) OnMouseMove() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	worldGrid := a.WorldGrid(&a.State().MouseWorld)

	for _, isector := range a.State().DB.AllOfType(core.SectorComponentIndex) {
		sector := isector.(*core.Sector)
		if sector.IsPointInside2D(worldGrid) {
			a.ContainingSector = sector
			break
		}
	}

	a.DetachFromSector()
	a.AttachToSector()
	if c := a.Components[core.BodyComponentIndex]; c != nil {
		body := c.(*core.Body)
		body.Pos.Original[0] = worldGrid[0]
		body.Pos.Original[1] = worldGrid[1]
		if a.ContainingSector != nil {
			floorZ, ceilZ := a.ContainingSector.SlopedZOriginal(worldGrid)
			body.Pos.Original[2] = (floorZ + ceilZ) / 2
		}
		body.Pos.Reset()
		a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
	}
}

func (a *AddEntity) OnMouseUp() {
	a.State().Modified = true
	a.ActionFinished(false, true, false)
}
func (a *AddEntity) Act() {
	a.Mode = "AddBody"
	a.SelectObjects(true, core.SelectableFromEntity(a.State().DB, a.Entity))
}
func (a *AddEntity) Cancel() {
	a.State().Lock.Lock()
	a.DetachFromSector()
	if a.Entity != 0 {
		a.State().DB.DetachAll(a.Entity)
	}
	a.State().Lock.Unlock()
	a.SelectObjects(true)
	a.ActionFinished(true, true, false)
}
func (a *AddEntity) Undo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	a.DetachFromSector()
	a.State().DB.DetachAll(a.Entity)
}
func (a *AddEntity) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	a.AttachToSector()
	a.AttachAll()
}

func (a *AddEntity) Frame() {}

func (a *AddEntity) Status() string {
	return ""
}
