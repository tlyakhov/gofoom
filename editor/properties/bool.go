package properties

import (
	"reflect"

	"github.com/gotk3/gotk3/gtk"
	"github.com/tlyakhov/gofoom/editor/actions"
)

func (g *Grid) fieldBool(index int, field *pgField) {
	origValue := false
	for _, v := range field.Values {
		origValue = origValue || v.Elem().Bool()
	}

	cb, _ := gtk.CheckButtonNew()
	cb.SetHExpand(true)
	cb.SetActive(origValue)
	cb.Connect("toggled", func(_ *gtk.CheckButton) {
		action := &actions.SetProperty{IEditor: g.IEditor, Fields: field.Values, ToSet: reflect.ValueOf(cb.GetActive())}
		g.NewAction(action)
		action.Act()
		origValue = cb.GetActive()
		g.Container.GrabFocus()
	})
	g.Container.Attach(cb, 2, index, 1, 1)
}
