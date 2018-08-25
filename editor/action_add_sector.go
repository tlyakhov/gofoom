package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/logic/provide"
)

type AddSectorAction struct {
	*Editor
	Sector core.AbstractSector
}

func (a *AddSectorAction) Act() {
	a.SetMapCursor("crosshair")
	a.State = "AddSector"
	a.SelectObjects([]concepts.ISerializable{a.Sector})
	//set cursor
}

func (a *AddSectorAction) Cancel() {
	a.RemoveFromMap()
	a.Sector.Physical().Segments = []*core.Segment{}
	a.SelectObjects([]concepts.ISerializable{})
	a.ActionFinished(true)
}

func (a *AddSectorAction) RemoveFromMap() {
	id := a.Sector.GetBase().ID
	if a.GameMap.Sectors[id] != nil {
		delete(a.GameMap.Sectors, id)
	}
}

func (a *AddSectorAction) AddToMap() {
	id := a.Sector.GetBase().ID
	a.Sector.Physical().Map = a.GameMap.Map
	a.GameMap.Sectors[id] = a.Sector
	provide.Passer.For(a.Sector).Recalculate()
}

func (a *AddSectorAction) OnMouseDown(button *gdk.EventButton) {
	a.State = "AddSectorSegment"

	seg := core.Segment{}
	seg.Initialize()
	seg.SetParent(a.Sector)
	seg.HiMaterial = a.GameMap.DefaultMaterial()
	seg.LoMaterial = a.GameMap.DefaultMaterial()
	seg.MidMaterial = a.GameMap.DefaultMaterial()
	seg.P = a.WorldGrid(a.MouseDownWorld)

	segs := a.Sector.Physical().Segments
	if len(segs) > 0 {
		seg.Prev = segs[len(segs)-1]
		seg.Next = segs[0]
		seg.Next.Prev = &seg
		seg.Prev.Next = &seg
	}

	a.Sector.Physical().Segments = append(segs, &seg)
	a.AddToMap()
}
func (a *AddSectorAction) OnMouseMove() {
	if a.State != "AddSectorSegment" {
		return
	}

	segs := a.Sector.Physical().Segments
	seg := segs[len(segs)-1]
	seg.P = a.WorldGrid(a.MouseWorld)
}

func (a *AddSectorAction) OnMouseUp() {
	a.State = "AddSector"

	segs := a.Sector.Physical().Segments
	if len(segs) > 1 {
		first := segs[0]
		last := segs[len(segs)-1]
		if last.P.Sub(first.P).Length() < SegmentSelectionEpsilon {
			a.Sector.Physical().Segments = segs[:(len(segs) - 1)]
			provide.Passer.For(a.Sector).Recalculate()
			a.ActionFinished(false)
		}
	}
	// TODO: right-mouse button end
}

func (a *AddSectorAction) Frame() {}

func (a *AddSectorAction) Undo() {
	a.RemoveFromMap()
}
func (a *AddSectorAction) Redo() {
	a.AddToMap()
}
