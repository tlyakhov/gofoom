package actions

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gdk"
)

type AddMob struct {
	state.IEditor

	Mode   string
	Mob    core.AbstractMob
	Sector core.AbstractSector
}

func (a *AddMob) RemoveFromMap() {
	phys := a.Mob.Physical()
	if phys.Sector != nil {
		delete(phys.Sector.Physical().Mobs, a.Mob.GetBase().Name)
	}
}

func (a *AddMob) AddToMap(sector core.AbstractSector) {
	a.Mob.Physical().Sector = sector
	a.Mob.Physical().Map = a.State().World.Map
	sector.Physical().Mobs[a.Mob.GetBase().Name] = a.Mob
	a.State().World.Recalculate()
}

func (a *AddMob) OnMouseDown(button *gdk.EventButton) {}

func (a *AddMob) OnMouseMove() {
	wg := a.WorldGrid(&a.State().MouseWorld)
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
	wg.To3D(&a.Mob.Physical().Pos.Original)
	floorZ, ceilZ := a.Sector.Physical().SlopedZOriginal(wg)
	a.Mob.Physical().Pos.Original[2] = (floorZ + ceilZ) / 2
	a.Mob.Physical().Pos.Reset()
}

func (a *AddMob) OnMouseUp() {
	a.State().Modified = true
	a.ActionFinished(false)
}
func (a *AddMob) Act() {
	a.Mode = "AddMob"
	a.SelectObjects([]concepts.ISerializable{a.Mob})
}
func (a *AddMob) Cancel() {
	a.RemoveFromMap()
	a.SelectObjects(nil)
	a.ActionFinished(true)
}
func (a *AddMob) Undo() {
	a.RemoveFromMap()
}
func (a *AddMob) Redo() {
	a.AddToMap(a.Sector)
}

func (a *AddMob) Frame() {}
