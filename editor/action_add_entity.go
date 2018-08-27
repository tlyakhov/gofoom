package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/core"
)

type AddEntityAction struct {
	*Editor
	Entity core.AbstractEntity
	Sector core.AbstractSector
}

func (a *AddEntityAction) RemoveFromMap() {
	phys := a.Entity.Physical()
	if phys.Sector != nil {
		delete(phys.Sector.Physical().Entities, a.Entity.GetBase().ID)
	}
}

func (a *AddEntityAction) AddToMap(sector core.AbstractSector) {
	a.Entity.Physical().Sector = sector
	a.Entity.Physical().Map = a.GameMap.Map
	sector.Physical().Entities[a.Entity.GetBase().ID] = a.Entity
	a.GameMap.Recalculate()
}

func (a *AddEntityAction) OnMouseDown(button *gdk.EventButton) {}

func (a *AddEntityAction) OnMouseMove() {
	wg := a.WorldGrid(a.MouseWorld)
	var sector core.AbstractSector

	for _, sector = range a.GameMap.Sectors {
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

func (a *AddEntityAction) OnMouseUp() {
	a.ActionFinished(false)
}
func (a *AddEntityAction) Act() {
	a.State = "AddEntity"
	a.SelectObjects([]concepts.ISerializable{a.Entity})
}
func (a *AddEntityAction) Cancel() {
	a.RemoveFromMap()
	a.SelectObjects(nil)
	a.ActionFinished(true)
}
func (a *AddEntityAction) Undo() {
	a.RemoveFromMap()
}
func (a *AddEntityAction) Redo() {
	a.AddToMap(a.Sector)
}

func (a *AddEntityAction) Frame() {}
