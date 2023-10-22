package actions

import (
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/core"

	"github.com/gotk3/gotk3/gdk"
)

type Delete struct {
	state.IEditor

	Selected []concepts.ISerializable
	Indices  map[concepts.ISerializable]int
}

func (a *Delete) Act() {
	a.Indices = make(map[concepts.ISerializable]int)
	a.Selected = make([]concepts.ISerializable, len(a.State().SelectedObjects))
	copy(a.Selected, a.State().SelectedObjects)

	for _, obj := range a.Selected {
		switch target := obj.(type) {
		case *state.MapPoint:
			for index, seg := range target.Sector.Physical().Segments {
				if seg == target.Segment {
					a.Indices[target] = index
				}
			}
		}
	}

	for _, obj := range a.Selected {
		switch target := obj.(type) {
		case *state.MapPoint:
			phys := target.Sector.Physical()
			indexToDelete := a.Indices[target]
			phys.Segments = append(phys.Segments[:indexToDelete], phys.Segments[indexToDelete+1:]...)
			for _, obj2 := range a.Selected {
				if mp2, ok := obj2.(*state.MapPoint); ok {
					if mp2.Sector.Physical() != phys {
						continue
					}
					if a.Indices[mp2] >= indexToDelete {
						a.Indices[mp2]--
					}
				}
			}
		case core.AbstractMob:
			if target == a.State().World.Player {
				// Otherwise weird things happen...
				continue
			}
			delete(target.GetSector().Physical().Mobs, target.GetBase().ID)
		case core.AbstractSector:
			delete(target.Physical().Map.Sectors, target.GetBase().ID)
		}
	}
	a.State().World.Recalculate()
	a.ActionFinished(false)
}
func (a *Delete) Cancel()                             {}
func (a *Delete) Frame()                              {}
func (a *Delete) OnMouseDown(button *gdk.EventButton) {}
func (a *Delete) OnMouseMove()                        {}
func (a *Delete) OnMouseUp()                          {}

func (a *Delete) Undo() {
}
func (a *Delete) Redo() {
}
