package actions

import (
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gdk"
)

type AddMaterial struct {
	state.IEditor
	core.Sampleable
}

func (a *AddMaterial) Act() {
	state := a.IEditor.State()
	name := a.GetBase().Name
	state.World.Materials[name] = a.Sampleable
	a.ActionFinished(false)
}
func (a *AddMaterial) Cancel() {
	a.ActionFinished(true)
}
func (a *AddMaterial) Undo() {
	state := a.IEditor.State()
	name := a.GetBase().Name
	delete(state.World.Materials, name)
}
func (a *AddMaterial) Redo() {
	a.Act()
}

func (a *AddMaterial) Frame()                              {}
func (a *AddMaterial) OnMouseDown(button *gdk.EventButton) {}
func (a *AddMaterial) OnMouseUp()                          {}
func (a *AddMaterial) OnMouseMove()                        {}
