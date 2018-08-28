package actions

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/editor/state"
)

type AddEntity struct {
	state.IEditor

	Mode   string
	Entity core.AbstractEntity
	Sector core.AbstractSector
}

func (a *AddEntity) RemoveFromMap() {
	phys := a.Entity.Physical()
	if phys.Sector != nil {
		delete(phys.Sector.Physical().Entities, a.Entity.GetBase().ID)
	}
}

func (a *AddEntity) AddToMap(sector core.AbstractSector) {
	a.Entity.Physical().Sector = sector
	a.Entity.Physical().Map = a.State().World.Map
	sector.Physical().Entities[a.Entity.GetBase().ID] = a.Entity
	a.State().World.Recalculate()
}

func (a *AddEntity) OnMouseDown(button *gdk.EventButton) {}

func (a *AddEntity) OnMouseMove() {
	wg := a.WorldGrid(a.State().MouseWorld)
	var sector core.AbstractSector

	for _, sector = range a.State().World.Sectors {
		if sector.IsPointInside2D(wg) {
			break
		}
	}

	if sector == nil {
		return
	}

	a.RemoveFromMap()
	a.AddToMap(sector)
	a.Sector = sector
	a.Entity.Physical().Pos = wg.To3D()
	floorZ, ceilZ := a.Sector.Physical().CalcFloorCeilingZ(wg)
	a.Entity.Physical().Pos.Z = (floorZ + ceilZ) / 2
}

func (a *AddEntity) OnMouseUp() {
	a.State().Modified = true
	a.ActionFinished(false)
}
func (a *AddEntity) Act() {
	a.Mode = "AddEntity"
	a.SelectObjects([]concepts.ISerializable{a.Entity})
}
func (a *AddEntity) Cancel() {
	a.RemoveFromMap()
	a.SelectObjects(nil)
	a.ActionFinished(true)
}
func (a *AddEntity) Undo() {
	a.RemoveFromMap()
}
func (a *AddEntity) Redo() {
	a.AddToMap(a.Sector)
}

func (a *AddEntity) Frame() {}
