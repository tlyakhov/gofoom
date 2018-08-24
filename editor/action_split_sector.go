package main

import (
	"github.com/tlyakhov/gofoom/concepts"

	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/core"
)

type SplitSectorAction struct {
	*Editor
	Splitters []*core.SectorSplitter
}

func (a *SplitSectorAction) OnMouseDown(button *gdk.EventButton) {}
func (a *SplitSectorAction) OnMouseMove()                        {}
func (a *SplitSectorAction) Frame()                              {}
func (a *SplitSectorAction) Act()                                {}

func (a *SplitSectorAction) Split(sector core.AbstractSector) {
	s := &core.SectorSplitter{
		Splitter1: a.WorldGrid(a.MouseDownWorld),
		Splitter2: a.WorldGrid(a.MouseWorld),
		Sector:    sector,
	}
	a.Splitters = append(a.Splitters, s)
	s.Do()
	delete(a.GameMap.Sectors, sector.GetBase().ID)
	for _, added := range s.Result {
		a.GameMap.Sectors[added.GetBase().ID] = added
	}
	a.GameMap.Recalculate()
}

func (a *SplitSectorAction) OnMouseUp() {
	a.Splitters = []*core.SectorSplitter{}

	// Split only selected if any, otherwise all sectors/segments.
	all := a.SelectedObjects
	if all == nil || len(all) == 0 || (len(all) == 1 && all[0] == a.GameMap.Map) {
		all = make([]concepts.ISerializable, len(a.GameMap.Sectors))
		i := 0
		for _, s := range a.GameMap.Sectors {
			all[i] = s
			i++
		}
	}

	for _, obj := range all {
		if sector, ok := obj.(core.AbstractSector); ok {
			a.Split(sector)
		} else if _, ok := obj.(MapPoint); ok {
			// TODO...
		}
	}
	a.ActionFinished(false)
}

func (a *SplitSectorAction) Cancel() {
	a.ActionFinished(true)
}

func (a *SplitSectorAction) Undo() {
}
func (a *SplitSectorAction) Redo() {
}
