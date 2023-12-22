package properties

import (
	"log"
	"reflect"

	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type StringLikeType interface {
	*concepts.Vector2 | *concepts.Vector3 | *concepts.Vector4 | *concepts.Matrix2

	String() string
}

func fieldStringLikeType[T StringLikeType](g *Grid, field *state.PropertyGridField) {
	origValue := ""
	for i, v := range field.Values {
		if i != 0 {
			origValue += ", "
		}
		currentValue := v.Interface().(T)
		origValue += currentValue.String()
	}

	entry := widget.NewEntry()
	entry.SetText(origValue)
	entry.OnChanged = func(text string) {
		var err error
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
			g.Focus(g.FContainer)
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
		g.Focus(g.FContainer)
	}
	cb := widget.NewCheck("Tweak", func(active bool) {
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
	g.FContainer.Add(container.NewBorder(nil, nil, nil, cb, entry))
}
