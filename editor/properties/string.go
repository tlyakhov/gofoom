package properties

import (
	"log"
	"reflect"

	"tlyakhov/gofoom/editor/actions"

	"github.com/gotk3/gotk3/gtk"
)

func (g *Grid) fieldString(index int, field *pgField) {
	origValue := ""
	i := 0
	for _, v := range field.Unique {
		if i != 0 {
			origValue += ", "
		}
		origValue += v.Elem().String()
		i++
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
		action := &actions.SetProperty{IEditor: g.IEditor, Fields: field.Values, ToSet: reflect.ValueOf(text)}
		g.NewAction(action)
		action.Act()
		origValue = text
		g.Container.GrabFocus()
	})
	g.Container.Attach(box, 2, index, 1, 1)
}
