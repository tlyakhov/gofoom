package properties

import (
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"
	"tlyakhov/gofoom/materials"

	"github.com/gotk3/gotk3/gtk"
)

func (g *Grid) fieldMaterials(index int, field *state.PropertyGridField) {
	if len(field.Values) > 1 {
		panic("Unexpectedly have multiple field values for a list of Materials.")
	}
	button, _ := gtk.ButtonNew()
	button.SetHExpand(true)
	button.SetLabel("Add Material")
	button.Connect("clicked", func(_ *gtk.Button) {
		mat := &materials.LitSampled{}
		mat.Initialize()
		action := &actions.AddMaterial{IEditor: g.IEditor, Sampleable: mat}
		g.NewAction(action)
		action.Act()
		g.Container.GrabFocus()
	})
	g.Container.Attach(button, 2, index, 1, 1)
}
