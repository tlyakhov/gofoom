// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"reflect"

	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/widget"
)

func (g *Grid) fieldBool(field *state.PropertyGridField) {
	origValue := false
	for _, v := range field.Values {
		origValue = origValue || v.Deref().Bool()
	}

	cb := gridAddOrUpdateWidgetAtIndex[*widget.Check](g)
	cb.OnChanged = nil
	cb.SetChecked(origValue)
	cb.OnChanged = func(active bool) {
		action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: reflect.ValueOf(active)}
		g.NewAction(action)
		action.Act()
		origValue = active
		g.Focus(g.GridWidget)
	}
}
