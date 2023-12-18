package properties

import (
	"log"
	"reflect"

	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"github.com/gotk3/gotk3/gtk"
)

type StringLikeType interface {
	*concepts.Vector2 | *concepts.Vector3 | *concepts.Vector4 | *concepts.Matrix2

	String() string
}

func fieldStringLikeType[T StringLikeType](g *Grid, index int, field *state.PropertyGridField) {
	origValue := ""
	for i, v := range field.Values {
		if i != 0 {
			origValue += ", "
		}
		currentValue := v.Interface().(T)
		origValue += currentValue.String()
	}

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
		var parsed any
		// This is a hack to switch on type of T
		var typedValue T
		switch any(typedValue).(type) {
		case *concepts.Vector2:
			parsed, err = concepts.ParseVector2(text)
		case *concepts.Vector3:
			parsed, err = concepts.ParseVector3(text)
		case *concepts.Vector4:
			parsed, err = concepts.ParseVector4(text)
		case *concepts.Matrix2:
			parsed, err = concepts.ParseMatrix2(text)
		}
		if err != nil {
			log.Printf("Couldn't parse %v from user entry. %v\n", reflect.TypeOf(typedValue).Name(), err)
			entry.SetText(origValue)
			g.Container.GrabFocus()
			return
		}
		currentValue := parsed.(T)
		action := &actions.SetProperty{
			IEditor:           g.IEditor,
			PropertyGridField: field,
			ToSet:             reflect.ValueOf(currentValue).Elem(),
		}
		g.NewAction(action)
		action.Act()
		origValue = currentValue.String()
		g.Container.GrabFocus()
	})
	g.Container.Attach(entry, 2, index, 1, 1)

	cb, _ := gtk.CheckButtonNew()
	cb.SetLabel("Tweak")
	cb.Connect("toggled", func(_ *gtk.CheckButton) {
		active := cb.GetActive()
		for i, v := range g.State().SelectedTransformables {
			if typed, ok := v.(T); ok && typed == field.Values[0].Interface() {
				if !active {
					s := g.State().SelectedTransformables
					g.State().SelectedTransformables = append(s[:i], s[i+1:]...)
				}
				return
			}
		}
		if active {
			g.State().SelectedTransformables = append(g.State().SelectedTransformables, field.Values[0].Interface())
		}
	})
	g.Container.Attach(cb, 3, index, 1, 1)
}
