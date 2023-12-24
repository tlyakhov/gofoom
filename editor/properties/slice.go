package properties

import (
	"fmt"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *Grid) fieldSlice(field *state.PropertyGridField) {
	// field.Type is *[]<something>
	elemType := field.Type.Elem().Elem()
	bLabel := fmt.Sprintf("Add %v", elemType.String())
	button := widget.NewButtonWithIcon(bLabel, theme.ContentAddIcon(), func() {
		action := &actions.AddSliceElement{IEditor: g.IEditor, SlicePtr: field.Values[0], Parent: field.Parent}
		g.NewAction(action)
		action.Act()
		g.Focus(g.FContainer)
	})
	g.FContainer.Add(button)
}
