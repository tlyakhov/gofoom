package actions

import (
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/core"

	"github.com/gotk3/gotk3/gdk"
)

type Delete struct {
	state.IEditor

	Selected []concepts.Attachable
	Indices  map[concepts.Attachable]int
}

func (a *Delete) Act() {
	a.Indices = make(map[concepts.Attachable]int)
	a.Selected = make([]concepts.Attachable, len(a.State().SelectedObjects))
	copy(a.Selected, a.State().SelectedObjects)

	for _, obj := range a.Selected {
		switch target := obj.(type) {
		case *state.MapPoint:
			for index, seg := range target.Sector.Segments {
				if seg == target.Segment {
					a.Indices[target] = index
				}
			}
		}
	}

	for _, obj := range a.Selected {
		switch target := obj.(type) {
		case *state.MapPoint:
			phys := target.Sector
			indexToDelete := a.Indices[target]
			phys.Segments = append(phys.Segments[:indexToDelete], phys.Segments[indexToDelete+1:]...)
			for _, obj2 := range a.Selected {
				if mp2, ok := obj2.(*state.MapPoint); ok {
					if mp2.Sector != phys {
						continue
					}
					if a.Indices[mp2] >= indexToDelete {
						a.Indices[mp2]--
					}
				}
			}
		case core.Mob:
			if target == a.State().World.Player {
				// Otherwise weird things happen...
				continue
			}
			delete(target.GetSector().Mobs, target.GetEntity().Name)
		case core.Sector:
			delete(target.Map.Sectors, target.GetEntity().Name)
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
