// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"log"
	"reflect"

	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type StringLikeType interface {
	concepts.Vector2 | concepts.Vector3 | concepts.Vector4 |
		*concepts.Vector2 | *concepts.Vector3 | *concepts.Vector4
}

func stringLikeTypeString[T StringLikeType, PT interface{ *T }](v PT) string {
	switch currentValue := any(v).(type) {
	case *concepts.Vector2:
		return currentValue.String()
	case *concepts.Vector3:
		return currentValue.String()
	case *concepts.Vector4:
		return currentValue.String()
	case **concepts.Vector2:
		return (*currentValue).String()
	case **concepts.Vector3:
		return (*currentValue).String()
	case **concepts.Vector4:
		return (*currentValue).String()
	}
	return ""
}

func fieldStringLikeType[T StringLikeType, PT interface{ *T }](g *Grid, field *state.PropertyGridField) {
	origValue := ""
	for i, v := range field.Values {
		if i != 0 {
			origValue += ", "
		}

		origValue += stringLikeTypeString(v.Value.Interface().(PT))
	}

	entry := widget.NewEntry()
	entry.SetText(origValue)
	entry.OnSubmitted = func(text string) {
		var err error
		var parsed any
		var currentValue PT
		switch any(currentValue).(type) {
		case *concepts.Vector2:
			parsed, err = concepts.ParseVector2(text)
		case *concepts.Vector3:
			parsed, err = concepts.ParseVector3(text)
		case *concepts.Vector4:
			parsed, err = concepts.ParseVector4(text)
		case **concepts.Vector2:
			parsed, err = concepts.ParseVector2(text)
		case **concepts.Vector3:
			parsed, err = concepts.ParseVector3(text)
		case **concepts.Vector4:
			parsed, err = concepts.ParseVector4(text)
		}
		if err != nil {
			log.Printf("Couldn't parse %v from user entry. %v\n", reflect.TypeOf(currentValue).Name(), err)
			entry.SetText(origValue)
			g.Focus(g.GridWidget)
			return
		}
		switch any(currentValue).(type) {
		case *concepts.Vector2:
			currentValue = parsed.(PT)
		case *concepts.Vector3:
			currentValue = parsed.(PT)
		case *concepts.Vector4:
			currentValue = parsed.(PT)
		case **concepts.Vector2:
			if v, ok := parsed.(T); ok {
				currentValue = &v
			}
		case **concepts.Vector3:
			if v, ok := parsed.(T); ok {
				currentValue = &v
			}
		case **concepts.Vector4:
			if v, ok := parsed.(T); ok {
				currentValue = &v
			}
		}
		g.ApplySetPropertyAction(field, reflect.ValueOf(currentValue).Elem())
		origValue = stringLikeTypeString(currentValue)
		g.Focus(g.GridWidget)
	}

	cb := widget.NewCheck("Tweak", nil)
	for _, v := range g.State().SelectedTransformables {
		if typed, ok := v.(PT); ok && typed == field.Values[0].Interface() {
			cb.SetChecked(true)
			break
		}
	}
	cb.OnChanged = func(active bool) {
		for i, v := range g.State().SelectedTransformables {
			if typed, ok := v.(PT); ok && typed == field.Values[0].Interface() {
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

	}

	c := gridAddOrUpdateWidgetAtIndex[*fyne.Container](g)
	c.Layout = layout.NewBorderLayout(nil, nil, nil, cb)
	c.Objects = []fyne.CanvasObject{entry, cb}
}
