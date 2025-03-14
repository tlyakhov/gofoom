// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"reflect"

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
		g.ApplySetPropertyAction(field, reflect.ValueOf(active))
		origValue = active
		g.Focus(g.GridWidget)
	}
	if field.Disabled() {
		cb.Disable()
	} else {
		cb.Enable()
	}
}
