// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"reflect"
	"strings"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (g *Grid) originalStrings(field *state.PropertyGridField) string {
	origValue := ""
	i := 0
	for _, v := range field.Unique {
		if i != 0 {
			origValue += ", "
		}
		e := v.Elem()
		if e.Kind() == reflect.String {
			origValue += e.String()
		} else {
			s := v.MethodByName("String").Call(nil)[0].String()
			origValue += s
		}
		i++
	}
	return origValue
}

func (g *Grid) fieldString(field *state.PropertyGridField, multiline bool) {
	origValue := g.originalStrings(field)

	var entry *widget.Entry

	if exp, ok := field.Values[0].Parent().(*core.Script); ok {
		label := widget.NewLabel("Compiled successfully")
		if exp.ErrorMessage != "" {
			label.Text = exp.ErrorMessage
			label.Importance = widget.DangerImportance
		} else {
			label.Importance = widget.SuccessImportance
		}
		entry = widget.NewEntry()
		c := gridAddOrUpdateWidgetAtIndex[*fyne.Container](g)
		c.Layout = layout.NewVBoxLayout()
		c.Objects = []fyne.CanvasObject{entry, label}
	} else {
		entry = gridAddOrUpdateWidgetAtIndex[*widget.Entry](g)
	}

	entry.MultiLine = multiline
	entry.OnSubmitted = nil
	entry.SetText(origValue)
	if _, ok := field.Values[0].Interface().(concepts.Set[string]); ok {
		entry.OnSubmitted = func(text string) {
			set := make(concepts.Set[string])
			split := strings.Split(text, ",")
			set.AddAll(split...)
			g.ApplySetPropertyAction(field, reflect.ValueOf(set))
		}
		return
	} else if _, ok := field.Values[0].Interface().(concepts.Set[ecs.Entity]); ok {
		entry.OnSubmitted = func(text string) {
			split := strings.Split(text, ",")
			g.ApplySetPropertyAction(field, reflect.ValueOf(ecs.DeserializeEntities(split)))
		}
		return
	}
	entry.OnSubmitted = func(text string) {
		g.ApplySetPropertyAction(field, reflect.ValueOf(text))
		origValue = text
	}
}
