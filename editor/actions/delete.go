package actions

import (
	"github.com/tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/core"
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
			phys.Segments = append(phys.Segments[:a.Indices[target]], phys.Segments[a.Indices[target]+1:]...)
		case core.AbstractEntity:
			if target == a.State().World.Player {
				// Otherwise weird things happen...
				continue
			}
			delete(target.GetSector().Physical().Entities, target.GetBase().ID)
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
