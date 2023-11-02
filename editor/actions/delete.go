package actions

import (
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"github.com/gotk3/gotk3/gdk"
)

type Delete struct {
	state.IEditor

	Selected []any
	Indices  map[*core.Segment]int
}

func (a *Delete) Act() {
	a.Indices = make(map[*core.Segment]int)
	a.Selected = make([]any, len(a.State().SelectedObjects))
	copy(a.Selected, a.State().SelectedObjects)

	for _, obj := range a.Selected {
		switch target := obj.(type) {
		case *core.Segment:
			for index, seg := range target.Sector.Segments {
				if seg == target {
					a.Indices[target] = index
				}
			}
		}
	}

	for _, obj := range a.Selected {
		switch target := obj.(type) {
		case *core.Segment:
			phys := target.Sector
			indexToDelete := a.Indices[target]
			phys.Segments = append(phys.Segments[:indexToDelete], phys.Segments[indexToDelete+1:]...)
			for _, obj2 := range a.Selected {
				if mp2, ok := obj2.(*core.Segment); ok {
					if mp2.Sector != phys {
						continue
					}
					if a.Indices[mp2] >= indexToDelete {
						a.Indices[mp2]--
					}
				}
			}
		case *concepts.EntityRef:
			if behaviors.PlayerFromDb(target) != nil {
				// Otherwise weird things happen...
				continue
			}
			a.State().DB.DetachAll(target.Entity)
			if body := core.BodyFromDb(target); body != nil {
				delete(body.Sector().Bodies, target.Entity)
			}
		}
	}
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
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
