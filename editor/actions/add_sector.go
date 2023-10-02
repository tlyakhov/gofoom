package actions

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/editor/state"
	"tlyakhov/gofoom/logic/provide"

	"github.com/gotk3/gotk3/gdk"
)

type AddSector struct {
	state.IEditor

	Mode   string
	Sector core.AbstractSector
}

func (a *AddSector) Act() {
	a.SetMapCursor("crosshair")
	a.Mode = "AddSector"
	a.SelectObjects([]concepts.ISerializable{a.Sector})
	//set cursor
}

func (a *AddSector) Cancel() {
	a.RemoveFromMap()
	a.Sector.Physical().Segments = []*core.Segment{}
	a.SelectObjects([]concepts.ISerializable{})
	a.ActionFinished(true)
}

func (a *AddSector) RemoveFromMap() {
	id := a.Sector.GetBase().ID
	if a.State().World.Sectors[id] != nil {
		delete(a.State().World.Sectors, id)
	}
}

func (a *AddSector) AddToMap() {
	id := a.Sector.GetBase().ID
	a.Sector.Physical().Map = a.State().World.Map
	a.State().World.Sectors[id] = a.Sector
	provide.Passer.For(a.Sector).Recalculate()
}

func (a *AddSector) OnMouseDown(button *gdk.EventButton) {
	a.Mode = "AddSectorSegment"

	seg := core.Segment{}
	seg.Initialize()
	seg.SetParent(a.Sector)
	seg.HiMaterial = a.State().World.DefaultMaterial()
	seg.LoMaterial = a.State().World.DefaultMaterial()
	seg.MidMaterial = a.State().World.DefaultMaterial()
	seg.P = a.WorldGrid(a.State().MouseDownWorld)

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
func (a *AddSector) OnMouseMove() {
	if a.Mode != "AddSectorSegment" {
		return
	}

	segs := a.Sector.Physical().Segments
	seg := segs[len(segs)-1]
	seg.P = a.WorldGrid(a.State().MouseWorld)
}

func (a *AddSector) OnMouseUp() {
	a.Mode = "AddSector"

	segs := a.Sector.Physical().Segments
	if len(segs) > 1 {
		first := segs[0]
		last := segs[len(segs)-1]
		if last.P.Sub(first.P).Length() < state.SegmentSelectionEpsilon {
			a.Sector.Physical().Segments = segs[:(len(segs) - 1)]
			provide.Passer.For(a.Sector).Recalculate()
			a.State().Modified = true
			a.ActionFinished(false)
		}
	}
	// TODO: right-mouse button end
}

func (a *AddSector) Frame() {}

func (a *AddSector) Undo() {
	a.RemoveFromMap()
}
func (a *AddSector) Redo() {
	a.AddToMap()
}
