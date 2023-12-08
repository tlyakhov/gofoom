package properties

import (
	"log"
	"reflect"
	"strconv"
	"strings"

	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gtk"
)

func (g *Grid) fieldNumber(index int, field *state.PropertyGridField) {
	origValue := ""
	for i, v := range field.Values {
		if i != 0 {
			origValue += ", "
		}
		if field.Type.String() == "*float64" {
			origValue += strconv.FormatFloat(v.Elem().Float(), 'f', -1, 64)
		} else if field.Type.String() == "*int" {
			origValue += strconv.Itoa(int(v.Elem().Int()))
		}
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
		var toSet reflect.Value
		if field.Type.String() == "*float64" {
			f, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
			if err != nil {
				log.Printf("Couldn't parse float64 from user entry. %v\n", err)
				box.SetText(origValue)
				g.Container.GrabFocus()
				return
			}
			toSet = reflect.ValueOf(f)
		} else {
			i, err := strconv.Atoi(strings.TrimSpace(text))
			if err != nil {
				log.Printf("Couldn't parse int from user entry. %v\n", err)
				box.SetText(origValue)
				g.Container.GrabFocus()
				return
			}
			toSet = reflect.ValueOf(i)
		}

		action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: toSet}
		g.NewAction(action)
		action.Act()
		origValue = text
		g.Container.GrabFocus()
	})
	g.Container.Attach(box, 2, index, 2, 1)
}
