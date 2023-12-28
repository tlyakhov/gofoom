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
		if action.RequiresLock() {
			g.State().Lock.Lock()
			action.Act()
			g.State().Lock.Unlock()
		} else {
			action.Act()
		}
		origValue = active
		g.Focus(g.FContainer)
	}
	g.FContainer.Add(cb)
}
