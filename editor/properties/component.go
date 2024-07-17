// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"fmt"
	"log"
	"reflect"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *Grid) fieldComponent(field *state.PropertyGridField) {
	// This will be a pointer
	parentType := ""
	entities := ""
	for _, v := range field.Values {
		entity := v.Interface().(concepts.Entity)
		if len(entities) > 0 {
			entities += ", "
		}
		entities += entity.Format()
		parentType = reflect.TypeOf(v.Parent()).Elem().String()
	}

	button := gridAddOrUpdateWidgetAtIndex[*widget.Button](g)
	button.Text = fmt.Sprintf("Remove %v from [%v]", parentType, entities)
	if len(button.Text) > 32 {
		button.Text = button.Text[:32] + "..."
	}
	button.Icon = theme.ContentRemoveIcon()
	button.OnTapped = func() {
		action := &actions.DeleteComponent{IEditor: g.IEditor, Components: make(map[concepts.Entity]concepts.Attachable)}
		for _, v := range field.Values {
			entity := v.Interface().(concepts.Entity)
			log.Printf("Detaching %v from %v", parentType, entity)
			action.Components[entity] = v.Parent().(concepts.Attachable)
		}
		g.NewAction(action)
		action.Act()
		g.Focus(g.GridWidget)
	}
}
