// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"log"
	"reflect"
	"strings"

	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type StringLikeType interface {
	concepts.Vector2 | concepts.Vector3 | concepts.Vector4 |
		*concepts.Vector2 | *concepts.Vector3 | *concepts.Vector4 |
		ecs.EntityTable | []ecs.Entity | containers.Set[ecs.ComponentID]
}

func stringLikeTypeString[T StringLikeType, PT interface{ *T }](v PT) string {
	switch currentValue := any(v).(type) {
	case *concepts.Vector2:
		return currentValue.String()
	case *concepts.Vector3:
		return currentValue.StringHuman(4)
	case *concepts.Vector4:
		return currentValue.String()
	case **concepts.Vector2:
		return (*currentValue).String()
	case **concepts.Vector3:
		return (*currentValue).String()
	case **concepts.Vector4:
		return (*currentValue).String()
	case *ecs.EntityTable:
		return currentValue.String()
	case *[]ecs.Entity:
		return (ecs.EntityTable)(*currentValue).String()
	case *containers.Set[ecs.ComponentID]:
		return ecs.SerializeComponentIDs(*currentValue)
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
		case *containers.Set[ecs.ComponentID]:
			ids := ecs.ParseComponentIDs(text)
			parsed = &ids
		case *ecs.EntityTable:
			entities := ecs.ParseEntityCSV(text)
			parsed = &entities
		case *[]ecs.Entity:
			split := strings.Split(text, ",")
			entities := make([]ecs.Entity, len(split))
			for i, s := range split {
				entity, subError := ecs.ParseEntity(s)
				if subError != nil {
					err = subError
					break
				}
				entities[i] = entity
			}
			parsed = &entities
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
		case *ecs.EntityTable:
			currentValue = parsed.(PT)
		case *[]ecs.Entity:
			currentValue = parsed.(PT)
		case *containers.Set[ecs.ComponentID]:
			currentValue = parsed.(PT)
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
