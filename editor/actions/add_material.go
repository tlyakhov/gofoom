package actions

import (
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gdk"
)

type AddMaterial struct {
	state.IEditor
	materials.Texturable
}

func (a *AddMaterial) Act() {
	state := a.IEditor.State()
	name := a.GetEntity().Name
	state.World.Materials[name] = a.Sampleable
	a.ActionFinished(false)
}
func (a *AddMaterial) Cancel() {
	a.ActionFinished(true)
}
func (a *AddMaterial) Undo() {
	state := a.IEditor.State()
	name := a.GetEntity().Name
	delete(state.World.Materials, name)
}
func (a *AddMaterial) Redo() {
	a.Act()
}

func (a *AddMaterial) Frame()                              {}
func (a *AddMaterial) OnMouseDown(button *gdk.EventButton) {}
func (a *AddMaterial) OnMouseUp()                          {}
func (a *AddMaterial) OnMouseMove()                        {}
