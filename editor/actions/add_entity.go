package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gdk"
)

type AddEntity struct {
	state.IEditor

	Mode             string
	EntityRef        *concepts.EntityRef
	Components       map[int]concepts.Attachable
	ContainingSector *core.Sector
}

func (a *AddEntity) RemoveFromMap() {
	if body := core.BodyFromDb(a.EntityRef); body != nil {
		if body.SectorEntityRef.Nil() {
			return
		}
		delete(body.Sector().Bodies, a.EntityRef.Entity)
	}
}

func (a *AddEntity) AttachAll() {
	for index, component := range a.Components {
		a.EntityRef.DB.Attach(index, a.EntityRef.Entity, component)
	}
}

func (a *AddEntity) AddToMap() {
	if c := a.Components[core.BodyComponentIndex]; c != nil {
		body := c.(*core.Body)
		if a.ContainingSector != nil {
			body.SectorEntityRef = a.ContainingSector.Ref()
			a.ContainingSector.Bodies[a.EntityRef.Entity] = *a.EntityRef
		}
		a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
	}
}

func (a *AddEntity) OnMouseDown(button *gdk.EventButton) {}

func (a *AddEntity) OnMouseMove() {
	worldGrid := a.WorldGrid(&a.State().MouseWorld)

	for _, isector := range a.State().DB.All(core.SectorComponentIndex) {
		sector := isector.(*core.Sector)
		if sector.IsPointInside2D(worldGrid) {
			a.ContainingSector = sector
			break
		}
	}

	a.RemoveFromMap()
	a.AddToMap()
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

func (a *AddEntity) OnMouseUp() {
	a.State().Modified = true
	a.ActionFinished(false)
}
func (a *AddEntity) Act() {
	a.Mode = "AddBody"
	a.SelectObjects([]any{a.EntityRef})
}
func (a *AddEntity) Cancel() {
	a.RemoveFromMap()
	a.EntityRef.DB.DetachAll(a.EntityRef.Entity)
	a.SelectObjects(nil)
	a.ActionFinished(true)
}
func (a *AddEntity) Undo() {
	a.RemoveFromMap()
	a.EntityRef.DB.DetachAll(a.EntityRef.Entity)
}
func (a *AddEntity) Redo() {
	a.AddToMap()
	a.AttachAll()
}

func (a *AddEntity) Frame() {}
