package main

import (
	"math"

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
	a.State = "AddSector"
	a.SelectObjects([]concepts.ISerializable{a.Sector})
	//set cursor
}

func (a *AddSectorAction) Cancel() {
	a.RemoveFromMap()
	a.SelectObjects([]concepts.ISerializable{})
	if index := len(a.UndoHistory) - 1; index >= 0 {
		a.UndoHistory = a.UndoHistory[:index]
	}
	a.ActionFinished()
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
	seg.A = a.MouseDownWorld

	if a.Grid.Visible {
		seg.A.X = math.Round(seg.A.X/GridSize) * GridSize
		seg.A.Y = math.Round(seg.A.Y/GridSize) * GridSize
	}
	seg.B = seg.A

	a.Sector.Physical().Segments = append(a.Sector.Physical().Segments, &seg)
	a.AddToMap()
}
func (a *AddSectorAction) OnMouseMove() {
	if a.State != "AddSectorSegment" {
		return
	}

	segs := a.Sector.Physical().Segments
	seg := segs[len(segs)-1]
	seg.A = a.MouseWorld

	if a.Grid.Visible {
		seg.A.X = math.Round(seg.A.X/GridSize) * GridSize
		seg.A.Y = math.Round(seg.A.Y/GridSize) * GridSize
	}
	seg.B = seg.A

	provide.Passer.For(a.Sector).Recalculate()
}
func (a *AddSectorAction) OnMouseUp() {
	a.State = "AddSector"

	segs := a.Sector.Physical().Segments
	if len(segs) > 1 {
		first := segs[0]
		last := segs[len(segs)-1]
		if last.A.Sub(first.A).Length() < SegmentSelectionEpsilon {
			a.Sector.Physical().Segments = segs[:(len(segs) - 1)]
			a.ActionFinished()
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
