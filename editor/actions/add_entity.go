package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gdk"
)

type AddEntity struct {
	state.IEditor

	Mode       string
	EntityRef  concepts.EntityRef
	Components map[int]concepts.Attachable
}

func (a *AddEntity) RemoveFromMap() {
	if mob := core.MobFromDb(&a.EntityRef); mob != nil {
		if mob.SectorEntityRef.Nil() {
			return
		}
		delete(mob.Sector().Mobs, a.EntityRef.Entity)
	}

}

func (a *AddEntity) AttachAll() {
	for index, component := range a.Components {
		a.EntityRef.DB.Attach(index, a.EntityRef.Entity, component)
	}
}

func (a *AddEntity) AddToMap(sector *core.Sector) {
	if c := a.Components[core.MobComponentIndex]; c != nil {
		mob := c.(*core.Mob)
		if sector != nil {
			mob.SectorEntityRef = sector.EntityRef()
			sector.Mobs[a.EntityRef.Entity] = a.EntityRef
		}
		a.State().DB.NewControllerSet().ActGlobal("Recalculate")
	}
}

func (a *AddEntity) OnMouseDown(button *gdk.EventButton) {}

func (a *AddEntity) OnMouseMove() {
	worldGrid := a.WorldGrid(&a.State().MouseWorld)
	var sector *core.Sector

	for _, isector := range a.State().DB.All(core.SectorComponentIndex) {
		sector = isector.(*core.Sector)
		if sector.IsPointInside2D(worldGrid) {
			break
		}
	}

	if sector == nil {
		return
	}

	a.RemoveFromMap()
	a.AddToMap(sector)
	if c := a.Components[core.MobComponentIndex]; c != nil {
		mob := c.(*core.Mob)
		mob.Pos.Original[0] = worldGrid[0]
		mob.Pos.Original[1] = worldGrid[1]
		if !mob.SectorEntityRef.Nil() {
			floorZ, ceilZ := sector.SlopedZOriginal(worldGrid)
			mob.Pos.Original[2] = (floorZ + ceilZ) / 2
		}
		mob.Pos.Reset()
		//a.State().World.Recalculate()
	}
}

func (a *AddEntity) OnMouseUp() {
	a.State().Modified = true
	a.ActionFinished(false)
}
func (a *AddEntity) Act() {
	a.Mode = "AddMob"
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
	a.AddToMap(a.Sector)
	a.AttachAll()
}

func (a *AddEntity) Frame() {}
