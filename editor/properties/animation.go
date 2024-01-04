package properties

import (
	"reflect"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *Grid) fieldAnimation(field *state.PropertyGridField) {
	origValue := field.Values[0].Elem()
	if origValue.IsNil() {
		button := widget.NewButtonWithIcon("Add Animation", theme.ContentAddIcon(), func() {
			parentValue := reflect.ValueOf(field.Parent)
			m := parentValue.MethodByName("NewAnimation")
			newAnimation := m.Call(nil)[0]
			action := &actions.SetProperty{
				IEditor:           g.IEditor,
				PropertyGridField: field,
				ToSet:             newAnimation,
			}
			g.NewAction(action)
			action.Act()
		})
		g.GridWidget.Objects = append(g.GridWidget.Objects, button)
	} else {
		button := widget.NewButtonWithIcon("Remove Animation", theme.ContentClearIcon(), func() {
			action := &actions.SetProperty{
				IEditor:           g.IEditor,
				PropertyGridField: field,
				ToSet:             reflect.Zero(origValue.Type()),
			}
			g.NewAction(action)
			action.Act()
		})
		g.GridWidget.Objects = append(g.GridWidget.Objects, button)
	}
}
