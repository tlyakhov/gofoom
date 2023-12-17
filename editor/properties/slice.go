package properties

import (
	"fmt"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gtk"
)

func (g *Grid) fieldSlice(index int, field *state.PropertyGridField) {
	// field.Type is *[]<something>
	elemType := field.Type.Elem().Elem()
	button, _ := gtk.ButtonNew()
	button.SetHExpand(true)
	button.SetLabel(fmt.Sprintf("Add %v", elemType.String()))
	button.Connect("clicked", func(_ *gtk.Button) {
		action := &actions.AddSliceElement{IEditor: g.IEditor, SlicePtr: field.Values[0], Parent: field.Parent}
		g.NewAction(action)
		action.Act()
		g.Container.GrabFocus()
	})
	g.Container.Attach(button, 2, index, 2, 1)
}
