package main

import (
	"reflect"

	"github.com/gotk3/gotk3/gtk"
)

func (e *Editor) PropertyGridFieldBool(index int, field *GridField) {
	origValue := false
	for _, v := range field.Values {
		origValue = origValue || v.Elem().Bool()
	}

	cb, _ := gtk.CheckButtonNew()
	cb.SetHExpand(true)
	cb.SetActive(origValue)
	cb.Connect("toggled", func(_ *gtk.CheckButton) {
		action := &SetPropertyAction{Editor: e, Fields: field.Values, ToSet: reflect.ValueOf(cb.GetActive())}
		e.NewAction(action)
		action.Act()
		origValue = cb.GetActive()
		e.PropertyGrid.GrabFocus()
	})
	e.PropertyGrid.Attach(cb, 2, index, 1, 1)
}
