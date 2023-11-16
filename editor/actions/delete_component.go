package actions

import (
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"github.com/gotk3/gotk3/gdk"
)

type DeleteComponent struct {
	state.IEditor

	Components map[uint64]concepts.Attachable
}

func (a *DeleteComponent) Act() {
	a.Redo()
	a.ActionFinished(false)
}
func (a *DeleteComponent) Cancel()                             {}
func (a *DeleteComponent) Frame()                              {}
func (a *DeleteComponent) OnMouseDown(button *gdk.EventButton) {}
func (a *DeleteComponent) OnMouseMove()                        {}
func (a *DeleteComponent) OnMouseUp()                          {}

func (a *DeleteComponent) Undo() {
	for entity, component := range a.Components {
		a.State().DB.AttachTyped(entity, component)
	}
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
}
func (a *DeleteComponent) Redo() {
	for _, component := range a.Components {
		a.State().DB.DetachByType(component)
	}
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
}
