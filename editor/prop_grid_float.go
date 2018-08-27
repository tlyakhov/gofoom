package main

import (
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func (e *Editor) PropertyGridFieldFloat64(index int, field *GridField) {
	origValue := ""
	for i, v := range field.Values {
		if i != 0 {
			origValue += ", "
		}
		origValue += strconv.FormatFloat(v.Elem().Float(), 'f', -1, 64)
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
		f, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
		if err != nil {
			log.Printf("Couldn't parse float64 from user entry. %v\n", err)
			box.SetText(origValue)
			e.PropertyGrid.GrabFocus()
			return
		}
		action := &SetPropertyAction{Editor: e, Fields: field.Values, ToSet: reflect.ValueOf(f)}
		e.NewAction(action)
		action.Act()
		origValue = text
		e.PropertyGrid.GrabFocus()
	})
	e.PropertyGrid.Attach(box, 2, index, 1, 1)
}
