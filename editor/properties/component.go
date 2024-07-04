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
	parentType := reflect.TypeOf(field.Parent)
	entityList := ""
	for _, v := range field.Values {
		entity := v.Elem().Interface().(concepts.Entity)
		if len(entityList) > 0 {
			entityList += ", "
		}
		entityList += entity.Format()
	}

	button := gridAddOrUpdateWidgetAtIndex[*widget.Button](g)
	button.Text = fmt.Sprintf("Remove %v from [%v]", parentType.Elem().String(), entityList)
	button.Icon = theme.ContentRemoveIcon()
	button.OnTapped = func() {
		action := &actions.DeleteComponent{IEditor: g.IEditor, Components: make(map[concepts.Entity]concepts.Attachable)}
		for _, v := range field.Values {
			entity := v.Elem().Interface().(concepts.Entity)
			log.Printf("Detaching %v from %v", parentType.String(), entity)
			action.Components[entity] = field.Parent.(concepts.Attachable)
		}
		g.NewAction(action)
		action.Act()
		g.Focus(g.GridWidget)
	}
}
