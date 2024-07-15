// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"log"
	"reflect"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/widget"
)

// This is actually not used anywhere, old way of handling animations
func (g *Grid) fieldAnimationTarget(field *state.PropertyGridField) {
	log.Printf("Someone called fieldAnimationTarget on %v, shouldn't happen.", field.Short())
	origValue := field.Values[0].Deref()
	opts := make([]string, 0)
	optsValues := make([]reflect.Value, 0)
	selectedIndex := -1

	// **concepts.SimVariable[float64] -> concepts.SimVariable[float64]
	searchType := field.Type.Elem().Elem()
	for _, component := range g.State().DB.AllComponents(field.Values[0].Entity) {
		if component == nil {
			continue
		}
		vComponent := reflect.ValueOf(component).Elem()
		for i := 0; i < vComponent.NumField(); i++ {
			target := vComponent.Field(i)
			name := reflect.TypeOf(component).Elem().Field(i).Name
			//			log.Printf("%v - %v (%v?)", name, target.Type().String(), field.Type.Elem().String())
			if target.Type() != searchType {
				continue
			}
			opts = append(opts, reflect.TypeOf(component).Elem().Name()+"."+name)
			optsValues = append(optsValues, target.Addr())
			if origValue.Interface() == target.Addr().Interface() {
				selectedIndex = len(optsValues) - 1
			}
		}
	}

	selectEntry := widget.NewSelect(opts, nil)
	if selectedIndex >= 0 {
		selectEntry.SetSelectedIndex(selectedIndex)
	}
	selectEntry.OnChanged = func(opt string) {
		g.ApplySetPropertyAction(field, optsValues[selectEntry.SelectedIndex()])
	}

	g.GridWidget.Objects = append(g.GridWidget.Objects, selectEntry)
}
