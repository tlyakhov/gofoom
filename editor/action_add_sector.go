package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/logic/provide"
)

type AddSectorAction struct {
	*Editor
	Sector           core.AbstractSector
	OriginalSegments []core.Segment
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
	seg.A = a.WorldGrid(a.MouseDownWorld)
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
	seg.A = a.WorldGrid(a.MouseWorld)
}

func (a *AddSectorAction) AutoPortal() {
	for _, sector := range a.GameMap.Sectors {
		if sector == a.Sector {
			continue
		}

		for _, segment := range sector.Physical().Segments {
			for _, addedSegment := range a.Sector.Physical().Segments {
				if addedSegment.Matches(segment) {
					a.OriginalSegments = append(a.OriginalSegments, *segment)
					addedSegment.AdjacentSector = sector
					addedSegment.AdjacentSegment = segment
					addedSegment.MidMaterial = nil
					segment.AdjacentSector = a.Sector
					segment.AdjacentSegment = addedSegment
					segment.MidMaterial = nil
				}
			}

		}
	}
	a.GameMap.ClearLightmaps()
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
			provide.Passer.For(a.Sector).Recalculate()
			a.AutoPortal()
			a.ActionFinished(false)
		}
	}
	// TODO: right-mouse button end
}

func (a *AddSectorAction) Frame() {}

func (a *AddSectorAction) Undo() {
	// Restore original segments.
	for _, segment := range a.OriginalSegments {
		segs := segment.Sector.Physical().Segments
		for index, link := range segs {
			if link.ID == segment.ID {
				segs[index] = &segment
			}
		}
	}
	a.OriginalSegments = []core.Segment{}
	a.RemoveFromMap()
}
func (a *AddSectorAction) Redo() {
	a.AddToMap()
	a.AutoPortal()
}
