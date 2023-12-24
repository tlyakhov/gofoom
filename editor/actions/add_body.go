package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/driver/desktop"
)

type AddBody struct {
	state.IEditor

	Mode             string
	EntityRef        *concepts.EntityRef
	Components       []concepts.Attachable
	ContainingSector *core.Sector
}

func (a *AddBody) DetachFromSector() {
	if body := core.BodyFromDb(a.EntityRef); body != nil {
		if body.SectorEntityRef.Nil() {
			return
		}
		delete(body.Sector().Bodies, a.EntityRef.Entity)
	}
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
}

func (a *AddBody) AttachAll() {
	for index, component := range a.Components {
		a.EntityRef.DB.Attach(index, a.EntityRef.Entity, component)
	}
}

func (a *AddBody) AttachToSector() {
	if c := a.Components[core.BodyComponentIndex]; c != nil {
		body := c.(*core.Body)
		if a.ContainingSector != nil {
			body.SectorEntityRef = a.ContainingSector.Ref()
			a.ContainingSector.Bodies[a.EntityRef.Entity] = a.EntityRef
		}
	}
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
}

func (a *AddBody) OnMouseDown(evt *desktop.MouseEvent) {}

func (a *AddBody) OnMouseMove() {
	worldGrid := a.WorldGrid(&a.State().MouseWorld)

	for _, isector := range a.State().DB.All(core.SectorComponentIndex) {
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
		a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
	}
}

func (a *AddBody) OnMouseUp() {
	a.State().Modified = true
	a.ActionFinished(false, true, false)
}
func (a *AddBody) Act() {
	a.Mode = "AddBody"
	a.SelectObjects([]any{a.EntityRef}, true)
}
func (a *AddBody) Cancel() {
	a.DetachFromSector()
	if a.EntityRef != nil {
		a.EntityRef.DB.DetachAll(a.EntityRef.Entity)
	}
	a.SelectObjects(nil, true)
	a.ActionFinished(true, true, false)
}
func (a *AddBody) Undo() {
	a.DetachFromSector()
	a.EntityRef.DB.DetachAll(a.EntityRef.Entity)
}
func (a *AddBody) Redo() {
	a.AttachToSector()
	a.AttachAll()
}

func (a *AddBody) Frame()             {}
func (a *AddBody) RequiresLock() bool { return true }
