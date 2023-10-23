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
	Original  []core.Sector
}

func (a *SplitSector) OnMouseDown(button *gdk.EventButton) {}
func (a *SplitSector) OnMouseMove()                        {}
func (a *SplitSector) Frame()                              {}
func (a *SplitSector) Act()                                {}

func (a *SplitSector) Split(sector core.Sector) {
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
	delete(a.State().World.Sectors, sector.GetEntity().Name)
	a.Original = append(a.Original, sector)
	for _, added := range s.Result {
		a.State().World.Sectors[added.GetEntity().Name] = added
	}
}

func (a *SplitSector) OnMouseUp() {
	a.Splitters = []*core.SectorSplitter{}

	// Split only selected if any, otherwise all sectors/segments.
	all := a.State().SelectedObjects
	if len(all) == 0 || (len(all) == 1 && all[0] == a.State().World.Map) {
		all = make([]concepts.Constructed, len(a.State().World.Sectors))
		i := 0
		for _, s := range a.State().World.Sectors {
			all[i] = s
			i++
		}
	}

	for _, obj := range all {
		if sector, ok := obj.(core.Sector); ok {
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
	mobs := []core.Mob{}

	for _, splitter := range a.Splitters {
		if splitter.Result == nil {
			continue
		}
		for _, added := range splitter.Result {
			for _, e := range added.Mobs {
				mobs = append(mobs, e)
				e.Sector = nil
			}
			added.Mobs = make(map[string]core.Mob)
			delete(a.State().World.Sectors, added.GetEntity().Name)
		}
	}
	for _, original := range a.Original {
		a.State().World.Sectors[original.GetEntity().Name] = original
		for _, e := range mobs {
			if original.IsPointInside2D(e.Pos.Original.To2D()) {
				original.Mobs[e.GetEntity().Name] = e
				e.SetParent(original)
			}
		}
	}
}
func (a *SplitSector) Redo() {
	mobs := []core.Mob{}

	for _, original := range a.Original {
		delete(a.State().World.Sectors, original.GetEntity().Name)
		for _, e := range original.Mobs {
			mobs = append(mobs, e)
			e.Sector = nil
		}
		original.Mobs = make(map[string]core.Mob)
	}

	for _, splitter := range a.Splitters {
		if splitter.Result == nil {
			continue
		}
		for _, added := range splitter.Result {
			a.State().World.Sectors[added.GetEntity().Name] = added
			for _, e := range mobs {
				if added.IsPointInside2D(e.Pos.Original.To2D()) {
					added.Mobs[e.GetEntity().Name] = e
					e.SetParent(added)
				}
			}
		}
	}

}
