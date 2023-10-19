package actions

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/core"

	"github.com/gotk3/gotk3/gdk"
)

type SplitSector struct {
	state.IEditor

	Splitters []*core.SectorSplitter
	Original  []core.AbstractSector
}

func (a *SplitSector) OnMouseDown(button *gdk.EventButton) {}
func (a *SplitSector) OnMouseMove()                        {}
func (a *SplitSector) Frame()                              {}
func (a *SplitSector) Act()                                {}

func (a *SplitSector) Split(sector core.AbstractSector) {
	s := &core.SectorSplitter{
		Splitter1: *a.WorldGrid(&a.State().MouseDownWorld),
		Splitter2: *a.WorldGrid(&a.State().MouseWorld),
		Sector:    sector,
	}
	a.Splitters = append(a.Splitters, s)
	s.Do()
	if s.Result == nil || len(s.Result) == 0 {
		return
	}
	delete(a.State().World.Sectors, sector.GetBase().ID)
	a.Original = append(a.Original, sector)
	for _, added := range s.Result {
		a.State().World.Sectors[added.GetBase().ID] = added
	}
}

func (a *SplitSector) OnMouseUp() {
	a.Splitters = []*core.SectorSplitter{}

	// Split only selected if any, otherwise all sectors/segments.
	all := a.State().SelectedObjects
	if len(all) == 0 || (len(all) == 1 && all[0] == a.State().World.Map) {
		all = make([]concepts.ISerializable, len(a.State().World.Sectors))
		i := 0
		for _, s := range a.State().World.Sectors {
			all[i] = s
			i++
		}
	}

	for _, obj := range all {
		if sector, ok := obj.(core.AbstractSector); ok {
			a.Split(sector)
		} else if mp, ok := obj.(state.MapPoint); ok {
			a.Split(mp.Segment.Sector)
		}
	}
	a.State().Modified = true
	a.ActionFinished(false)
}

func (a *SplitSector) Cancel() {
	a.ActionFinished(true)
}

func (a *SplitSector) Undo() {
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
			delete(a.State().World.Sectors, added.GetBase().ID)
		}
	}
	for _, original := range a.Original {
		a.State().World.Sectors[original.GetBase().ID] = original
		for _, e := range entities {
			if original.Physical().IsPointInside2D(e.Physical().Pos.Original.To2D()) {
				original.Physical().Entities[e.GetBase().ID] = e
				e.SetParent(original)
			}
		}
	}
}
func (a *SplitSector) Redo() {
	entities := []core.AbstractEntity{}

	for _, original := range a.Original {
		delete(a.State().World.Sectors, original.GetBase().ID)
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
			a.State().World.Sectors[added.GetBase().ID] = added
			for _, e := range entities {
				if added.Physical().IsPointInside2D(e.Physical().Pos.Original.To2D()) {
					added.Physical().Entities[e.GetBase().ID] = e
					e.SetParent(added)
				}
			}
		}
	}

}
