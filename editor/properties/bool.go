package properties

import (
	"reflect"

	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/widget"
)

func (g *Grid) fieldBool(field *state.PropertyGridField) {
	origValue := false
	for _, v := range field.Values {
		origValue = origValue || v.Elem().Bool()
	}

	cb := widget.NewCheck("", nil)
	cb.SetChecked(origValue)
	cb.OnChanged = func(active bool) {
		action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: reflect.ValueOf(active)}
		g.NewAction(action)
		action.Act()
		origValue = active
		g.Focus(g.GridWidget)
	}
	g.GridWidget.Objects = append(g.GridWidget.Objects, cb)

}
