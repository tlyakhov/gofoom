package properties

import (
	"log"
	"reflect"

	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"github.com/gotk3/gotk3/gtk"
)

func (g *Grid) fieldVector2(index int, field *state.PropertyGridField) {
	origValue := ""
	for i, v := range field.Values {
		if i != 0 {
			origValue += ", "
		}
		vec := v.Elem().Interface().(concepts.Vector2)
		origValue += vec.String()
	}

	box, _ := gtk.EntryNew()
	box.SetHExpand(true)
	box.SetText(origValue)
	box.Connect("activate", func(_ *gtk.Entry) {
		text, err := box.GetText()
		if err != nil {
			log.Printf("Couldn't get text from gtk.Entry. %v\n", err)
			box.SetText(origValue)
			g.Container.GrabFocus()
			return
		}
		vec, err := concepts.ParseVector2(text)
		if err != nil {
			log.Printf("Couldn't parse Vector2 from user entry. %v\n", err)
			box.SetText(origValue)
			g.Container.GrabFocus()
			return
		}
		action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: reflect.ValueOf(vec).Elem()}
		g.NewAction(action)
		action.Act()
		origValue = vec.String()
		g.Container.GrabFocus()
	})
	g.Container.Attach(box, 2, index, 2, 1)
}
