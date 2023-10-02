package properties

import (
	"log"
	"reflect"

	"tlyakhov/gofoom/editor/actions"

	"tlyakhov/gofoom/concepts"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func (g *Grid) fieldVector2(index int, field *pgField) {
	origValue := ""
	for i, v := range field.Values {
		if i != 0 {
			origValue += ", "
		}
		origValue += v.Elem().Interface().(concepts.Vector2).String()
	}

	box, _ := gtk.EntryNew()
	box.SetHExpand(true)
	box.SetText(origValue)
	box.Connect("activate", func(_ *glib.Object) {
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
		action := &actions.SetProperty{IEditor: g.IEditor, Fields: field.Values, ToSet: reflect.ValueOf(vec)}
		g.NewAction(action)
		action.Act()
		origValue = vec.String()
		g.Container.GrabFocus()
	})
	g.Container.Attach(box, 2, index, 1, 1)
}
