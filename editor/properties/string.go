// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"reflect"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/actions"
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
		origValue += v.Elem().String()
		i++
	}
	return origValue
}

func (g *Grid) fieldString(field *state.PropertyGridField, multiline bool) {
	origValue := g.originalStrings(field)

	var entry *widget.Entry

	if exp, ok := field.Parent.(*core.Script); ok {
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
	entry.OnSubmitted = func(text string) {
		action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: reflect.ValueOf(text)}
		g.NewAction(action)
		action.Act()
		origValue = text
	}
}
