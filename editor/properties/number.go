// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"log"
	"reflect"
	"strconv"
	"strings"

	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/widget"
)

func (g *Grid) fieldNumber(field *state.PropertyGridField) {
	origValue := ""
	for i, v := range field.Values {
		if i != 0 {
			origValue += ", "
		}
		if field.Type.String() == "*float64" {
			origValue += strconv.FormatFloat(v.Deref().Float(), 'f', -1, 64)
		} else if field.Type.String() == "*int" {
			origValue += strconv.Itoa(int(v.Deref().Int()))
		}
	}

	entry := gridAddOrUpdateWidgetAtIndex[*widget.Entry](g)
	entry.OnSubmitted = nil
	entry.SetText(origValue)
	entry.OnSubmitted = func(text string) {
		var toSet reflect.Value
		if field.Type.String() == "*float64" {
			f, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
			if err != nil {
				log.Printf("Couldn't parse float64 from user entry. %v\n", err)
				entry.SetText(origValue)
				return
			}
			toSet = reflect.ValueOf(f)
		} else {
			i, err := strconv.Atoi(strings.TrimSpace(text))
			if err != nil {
				log.Printf("Couldn't parse int from user entry. %v\n", err)
				entry.SetText(origValue)
				return
			}
			toSet = reflect.ValueOf(i)
		}

		action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: toSet}
		g.NewAction(action)
		action.Act()
		origValue = text
	}
}
