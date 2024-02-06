package actions

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/components/core"

	"fyne.io/fyne/v2/driver/desktop"
)

type SplitSector struct {
	state.IEditor

	Splitters []*controllers.SectorSplitter
	Original  map[uint64][]concepts.Attachable
}

func (a *SplitSector) OnMouseDown(evt *desktop.MouseEvent) {}
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
			if component == nil {
				continue
			}
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

	visited := make(map[uint64]bool)
	for _, obj := range all {
		switch target := obj.(type) {
		case *concepts.EntityRef:
			if _, ok := visited[target.Entity]; ok {
				continue
			}
			if sector := core.SectorFromDb(target); sector != nil {
				a.Split(sector)
			}
			visited[target.Entity] = true
		case *core.SectorSegment:
			a.Split(target.Sector)
			visited[target.Sector.Entity] = true
		}
	}
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}

func (a *SplitSector) Cancel() {
	a.ActionFinished(true, true, true)
}

func (a *SplitSector) Undo() {
	bodies := make([]*concepts.EntityRef, 0)

	for _, splitter := range a.Splitters {
		if splitter.Result == nil {
			continue
		}
		for _, addedComponents := range splitter.Result {
			sector := addedComponents[core.SectorComponentIndex].(*core.Sector)
			for _, ibody := range sector.Bodies {
				bodies = append(bodies, ibody)
				if body := core.BodyFromDb(ibody); body != nil {
					body.SectorEntityRef = nil
				}
			}
			sector.Bodies = make(map[uint64]*concepts.EntityRef)
			a.State().DB.DetachAll(sector.Entity)
		}
	}
	for entity, originalComponents := range a.Original {
		for index, component := range originalComponents {
			if component == nil {
				continue
			}
			a.State().DB.Attach(index, entity, component)
			if sector, ok := component.(*core.Sector); ok {
				for _, ibody := range bodies {
					if body := core.BodyFromDb(ibody); body != nil {
						if sector.IsPointInside2D(body.Pos.Original.To2D()) {
							body.SectorEntityRef = sector.EntityRef
							sector.Bodies[ibody.Entity] = body.EntityRef
						}

					}
				}
			}
		}
	}
}
func (a *SplitSector) Redo() {
	bodies := make([]*concepts.EntityRef, 0)

	for entity, originalComponents := range a.Original {
		sector := originalComponents[core.SectorComponentIndex].(*core.Sector)
		for _, ibody := range sector.Bodies {
			bodies = append(bodies, ibody)
			if body := core.BodyFromDb(ibody); body != nil {
				body.SectorEntityRef = nil
			}
		}
		sector.Bodies = make(map[uint64]*concepts.EntityRef)
		a.State().DB.DetachAll(entity)
	}

	for _, splitter := range a.Splitters {
		if splitter.Result == nil {
			continue
		}
		for _, addedComponents := range splitter.Result {
			for index, component := range addedComponents {
				if component == nil {
					continue
				}
				a.State().DB.Attach(index, component.Ref().Entity, component)
				if sector, ok := component.(*core.Sector); ok {
					for _, ibody := range bodies {
						if body := core.BodyFromDb(ibody); body != nil {
							if sector.IsPointInside2D(body.Pos.Original.To2D()) {
								body.SectorEntityRef = sector.EntityRef
								sector.Bodies[ibody.Entity] = body.EntityRef
							}

						}
					}
				}
			}
		}
	}

}
