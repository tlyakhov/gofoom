// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"reflect"
	"slices"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *Grid) fieldAnimation(field *state.PropertyGridField) {
	origValue := field.Values[0].Deref()
	if origValue.IsNil() {
		button := gridAddOrUpdateWidgetAtIndex[*widget.Button](g)
		button.Text = "Add Animation"
		button.Icon = theme.ContentAddIcon()
		button.OnTapped = func() {
			parentValue := reflect.ValueOf(field.Values[0].Parent)
			m := parentValue.MethodByName("NewAnimation")
			newAnimation := m.Call(nil)[0]
			g.ApplySetPropertyAction(field, newAnimation)
		}
	} else {
		button := gridAddOrUpdateWidgetAtIndex[*widget.Button](g)
		button.Text = "Remove Animation"
		button.Icon = theme.ContentClearIcon()
		button.OnTapped = func() {
			g.ApplySetPropertyAction(field, reflect.Zero(origValue.Type()))
		}
	}
}

func (g *Grid) fieldTweeningFunc(field *state.PropertyGridField) {
	origValue := field.Values[0].Deref().Pointer()

	opts := make([]string, 0)
	for name := range concepts.TweeningFuncs {
		opts = append(opts, name)
	}
	slices.Sort(opts)
	optValues := make([]reflect.Value, 0)
	selectedIndex := 0
	for i, name := range opts {
		f := reflect.ValueOf(concepts.TweeningFuncs[name])
		optValues = append(optValues, f)
		if f.Pointer() == origValue {
			selectedIndex = i
		}
	}

	s := gridAddOrUpdateWidgetAtIndex[*widget.Select](g)
	s.Options = opts
	s.OnChanged = nil
	s.SetSelectedIndex(selectedIndex)
	s.PlaceHolder = "Select tweening function"
	s.OnChanged = func(opt string) {
		g.ApplySetPropertyAction(field, optValues[s.SelectedIndex()])
	}
}
