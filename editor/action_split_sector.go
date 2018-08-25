package main

import (
	"github.com/tlyakhov/gofoom/concepts"

	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/core"
)

type SplitSectorAction struct {
	*Editor
	Splitters []*core.SectorSplitter
	Original  []core.AbstractSector
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
	if s.Result == nil || len(s.Result) == 0 {
		return
	}
	delete(a.GameMap.Sectors, sector.GetBase().ID)
	a.Original = append(a.Original, sector)
	for _, added := range s.Result {
		a.GameMap.Sectors[added.GetBase().ID] = added
	}
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
		} else if mp, ok := obj.(MapPoint); ok {
			a.Split(mp.Segment.Sector)
		}
	}
	a.ActionFinished(false)
}

func (a *SplitSectorAction) Cancel() {
	a.ActionFinished(true)
}

func (a *SplitSectorAction) Undo() {
	entities := []core.AbstractEntity{}

	for _, splitter := range a.Splitters {
		if splitter.Result == nil {
			continue
		}
		for _, added := range splitter.Result {
			for _, e := range added.Physical().Entities {
				entities = append(entities, e)
				e.Physical().Sector = nil
			}
			added.Physical().Entities = make(map[string]core.AbstractEntity)
			delete(a.GameMap.Sectors, added.GetBase().ID)
		}
	}
	for _, original := range a.Original {
		a.GameMap.Sectors[original.GetBase().ID] = original
		for _, e := range entities {
			if original.Physical().IsPointInside2D(e.Physical().Pos.To2D()) {
				original.Physical().Entities[e.GetBase().ID] = e
				e.SetParent(original)
			}
		}
	}
}
func (a *SplitSectorAction) Redo() {
	entities := []core.AbstractEntity{}

	for _, original := range a.Original {
		delete(a.GameMap.Sectors, original.GetBase().ID)
		for _, e := range original.Physical().Entities {
			entities = append(entities, e)
			e.Physical().Sector = nil
		}
		original.Physical().Entities = make(map[string]core.AbstractEntity)
	}

	for _, splitter := range a.Splitters {
		if splitter.Result == nil {
			continue
		}
		for _, added := range splitter.Result {
			a.GameMap.Sectors[added.GetBase().ID] = added
			for _, e := range entities {
				if added.Physical().IsPointInside2D(e.Physical().Pos.To2D()) {
					added.Physical().Entities[e.GetBase().ID] = e
					e.SetParent(added)
				}
			}
		}
	}

}
