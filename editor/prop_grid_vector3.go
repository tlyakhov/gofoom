package main

import (
	"log"
	"reflect"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/tlyakhov/gofoom/concepts"
)

func (e *Editor) PropertyGridFieldVector3(index int, field *GridField) {
	origValue := ""
	for i, v := range field.Values {
		if i != 0 {
			origValue += ", "
		}
		origValue += v.Elem().Interface().(concepts.Vector3).String()
	}

	box, _ := gtk.EntryNew()
	box.SetHExpand(true)
	box.SetText(origValue)
	box.Connect("activate", func(_ *glib.Object) {
		text, err := box.GetText()
		if err != nil {
			log.Printf("Couldn't get text from gtk.Entry. %v\n", err)
			box.SetText(origValue)
			e.PropertyGrid.GrabFocus()
			return
		}
		vec, err := concepts.ParseVector3(text)
		if err != nil {
			log.Printf("Couldn't parse Vector3 from user entry. %v\n", err)
			box.SetText(origValue)
			e.PropertyGrid.GrabFocus()
			return
		}
		action := &SetPropertyAction{Editor: e, Fields: field.Values, ToSet: reflect.ValueOf(vec)}
		e.NewAction(action)
		action.Act()
		origValue = vec.String()
		e.PropertyGrid.GrabFocus()
	})
	e.PropertyGrid.Attach(box, 2, index, 1, 1)
}
