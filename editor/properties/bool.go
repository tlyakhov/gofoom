package properties

import (
	"reflect"

	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gtk"
)

func (g *Grid) fieldBool(index int, field *state.PropertyGridField) {
	origValue := false
	for _, v := range field.Values {
		origValue = origValue || v.Elem().Bool()
	}

	cb, _ := gtk.CheckButtonNew()
	cb.SetHExpand(true)
	cb.SetActive(origValue)
	cb.Connect("toggled", func(_ *gtk.CheckButton) {
		action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: reflect.ValueOf(cb.GetActive())}
		g.NewAction(action)
		action.Act()
		origValue = cb.GetActive()
		g.Container.GrabFocus()
	})
	g.Container.Attach(cb, 2, index, 1, 1)
}
