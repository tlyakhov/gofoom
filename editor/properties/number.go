// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"log"
	"reflect"
	"strconv"
	"strings"

	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/widget"
)

func (g *Grid) fieldNumber(field *state.PropertyGridField) {
	origValue := ""
	for i, v := range field.Values {
		if i != 0 {
			origValue += ", "
		}
		switch v.Value.Interface().(type) {
		case *float32:
			origValue += strconv.FormatFloat(v.Deref().Float(), 'f', -1, 32)
		case *float64:
			origValue += strconv.FormatFloat(v.Deref().Float(), 'f', -1, 64)
		case *int:
			origValue += strconv.FormatInt(v.Deref().Int(), 10)
		case *uint32:
			origValue += strconv.FormatUint(v.Deref().Uint(), 10)
		case *uint64:
			origValue += strconv.FormatUint(v.Deref().Uint(), 10)
		}
	}

	entry := gridAddOrUpdateWidgetAtIndex[*widget.Entry](g)
	if field.Disabled() {
		entry.Disable()
	} else {
		entry.Enable()
	}

	entry.OnSubmitted = nil
	entry.SetText(origValue)
	entry.OnSubmitted = func(text string) {
		var toSet reflect.Value
		switch field.Type.String() {
		case "*float32":
			f, err := strconv.ParseFloat(strings.TrimSpace(text), 32)
			if err != nil {
				log.Printf("Couldn't parse float32 from user entry. %v\n", err)
				entry.SetText(origValue)
				return
			}
			toSet = reflect.ValueOf(float32(f))
		case "*float64":
			f, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
			if err != nil {
				log.Printf("Couldn't parse float64 from user entry. %v\n", err)
				entry.SetText(origValue)
				return
			}
			toSet = reflect.ValueOf(f)
		case "*int":
			i, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
			if err != nil {
				log.Printf("Couldn't parse int from user entry. %v\n", err)
				entry.SetText(origValue)
				return
			}
			toSet = reflect.ValueOf(int(i))
		case "*uint32":
			i, err := strconv.ParseUint(strings.TrimSpace(text), 10, 32)
			if err != nil {
				log.Printf("Couldn't parse uint32 from user entry. %v\n", err)
				entry.SetText(origValue)
				return
			}
			toSet = reflect.ValueOf(uint32(i))
		case "*uint64":
			i, err := strconv.ParseUint(strings.TrimSpace(text), 10, 64)
			if err != nil {
				log.Printf("Couldn't parse uint64 from user entry. %v\n", err)
				entry.SetText(origValue)
				return
			}
			toSet = reflect.ValueOf(i)
		}
		g.ApplySetPropertyAction(field, toSet)
		origValue = text
	}
}
