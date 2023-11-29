package properties

import (
	"log"
	"reflect"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gtk"
)

func (g *Grid) originalStrings(field *state.PropertyGridField) string {
	origValue := ""
	i := 0
	for _, v := range field.Unique {
		if i != 0 {
			origValue += ", "
		}
		origValue += v.Elem().String()
		i++
	}
	return origValue
}

func (g *Grid) fieldString(index int, field *state.PropertyGridField) {
	origValue := g.originalStrings(field)

	entry, _ := gtk.EntryNew()
	entry.SetHExpand(true)
	entry.SetText(origValue)
	entry.Connect("activate", func(_ *gtk.Entry) {
		text, err := entry.GetText()
		if err != nil {
			log.Printf("Couldn't get text from gtk.Entry. %v\n", err)
			entry.SetText(origValue)
			g.Container.GrabFocus()
			return
		}
		action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: reflect.ValueOf(text)}
		g.NewAction(action)
		action.Act()
		origValue = text
		g.Container.GrabFocus()
	})

	if exp, ok := field.Parent.(*core.Expression); ok {
		if exp.ErrorMessage != "" {
			entry.SetTooltipText(exp.ErrorMessage)
		} else {
			entry.SetTooltipText("Success")
		}
	}

	g.Container.Attach(entry, 2, index, 2, 1)
}
