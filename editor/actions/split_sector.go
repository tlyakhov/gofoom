package actions

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/components/core"

	"github.com/gotk3/gotk3/gdk"
)

type SplitSector struct {
	state.IEditor

	Splitters []*controllers.SectorSplitter
	Original  map[uint64][]concepts.Attachable
}

func (a *SplitSector) OnMouseDown(button *gdk.EventButton) {}
func (a *SplitSector) OnMouseMove()                        {}
func (a *SplitSector) Frame()                              {}
func (a *SplitSector) Act()                                {}

func (a *SplitSector) Split(sector *core.Sector) {
	s := &controllers.SectorSplitter{
		Splitter1: *a.WorldGrid(&a.State().MouseDownWorld),
		Splitter2: *a.WorldGrid(&a.State().MouseWorld),
		Sector:    sector,
	}
	a.Splitters = append(a.Splitters, s)
	s.Do()
	if s.Result == nil || len(s.Result) == 0 {
		return
	}
	// Copy original sector's components to preserve them
	copy(a.Original[sector.Entity], sector.DB.EntityComponents[sector.Entity])
	// Detach the original from the DB
	sector.DB.DetachAll(sector.Entity)
	// Attach the cloned entities/components
	for _, added := range s.Result {
		for index, component := range added {
			a.State().DB.Attach(index, component.Ref().Entity, component)
		}
	}
}

func (a *SplitSector) OnMouseUp() {
	a.Splitters = []*controllers.SectorSplitter{}

	// Split only selected if any, otherwise all sectors/segments.
	all := a.State().SelectedObjects
	if len(all) == 0 || (len(all) == 1 && all[0] == a.State().DB) {
		allSectors := a.State().DB.Components[core.SectorComponentIndex]
		all = make([]any, len(allSectors))
		i := 0
		for _, s := range allSectors {
			all[i] = s.Ref()
			i++
		}
	}

	for _, obj := range all {
		switch target := obj.(type) {
		case *concepts.EntityRef:
			if sector := core.SectorFromDb(target); sector != nil {
				a.Split(sector)
			}
		case *core.Segment:
			a.Split(target.Sector)
		}
	}
	a.State().Modified = true
	a.ActionFinished(false)
}

func (a *SplitSector) Cancel() {
	a.ActionFinished(true)
}

func (a *SplitSector) Undo() {
	bodys := []core.Body{}

	for _, splitter := range a.Splitters {
		if splitter.Result == nil {
			continue
		}
		for _, added := range splitter.Result {
			for _, e := range added.Bodies {
				bodys = append(bodys, e)
				e.Sector = nil
			}
			added.Bodies = make(map[string]core.Body)
			delete(a.State().World.Sectors, added.GetEntity().Name)
		}
	}
	for _, original := range a.Original {
		a.State().World.Sectors[original.GetEntity().Name] = original
		for _, e := range bodys {
			if original.IsPointInside2D(e.Pos.Original.To2D()) {
				original.Bodys[e.GetEntity().Name] = e
				e.SetParent(original)
			}
		}
	}
}
func (a *SplitSector) Redo() {
	bodys := []core.Body{}

	for _, original := range a.Original {
		delete(a.State().World.Sectors, original.GetEntity().Name)
		for _, e := range original.Bodys {
			bodys = append(bodys, e)
			e.Sector = nil
		}
		original.Bodys = make(map[string]core.Body)
	}

	for _, splitter := range a.Splitters {
		if splitter.Result == nil {
			continue
		}
		for _, added := range splitter.Result {
			a.State().World.Sectors[added.GetEntity().Name] = added
			for _, e := range bodys {
				if added.IsPointInside2D(e.Pos.Original.To2D()) {
					added.Bodys[e.GetEntity().Name] = e
					e.SetParent(added)
				}
			}
		}
	}

}
