// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"reflect"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *Grid) fieldSliceAdd(field *state.PropertyGridField, concreteType reflect.Type) {
	action := &actions.AddSliceElement{IEditor: g.IEditor, SlicePtr: field.Values[0], Parent: field.Parent, Concrete: concreteType}
	g.NewAction(action)
	action.Act()
	g.Focus(g.GridWidget)
}

var animationTypes = map[string]reflect.Type{
	"Animation[int]":     concepts.ReflectType[concepts.Animation[int]](),
	"Animation[float64]": concepts.ReflectType[concepts.Animation[float64]](),
	"Animation[Vector2]": concepts.ReflectType[concepts.Animation[concepts.Vector2]](),
	"Animation[Vector3]": concepts.ReflectType[concepts.Animation[concepts.Vector3]](),
	"Animation[Vector4]": concepts.ReflectType[concepts.Animation[concepts.Vector4]](),
}

func (g *Grid) fieldSlice(field *state.PropertyGridField) {
	// field.Type is *[]<something>
	elemType := field.Type.Elem().Elem()
	if elemType == concepts.ReflectType[concepts.Animated]() {
		buttons := make([]fyne.CanvasObject, len(animationTypes))
		i := 0
		for name, t := range animationTypes {
			t := t // To ensure correct scope for closure
			buttons[i] = widget.NewButtonWithIcon("Add "+name, theme.ContentAddIcon(), func() { g.fieldSliceAdd(field, t) })
			i++
		}
		c := gridAddOrUpdateWidgetAtIndex[*fyne.Container](g)
		c.Layout = layout.NewVBoxLayout()
		c.Objects = buttons
	} else {
		button := gridAddOrUpdateWidgetAtIndex[*widget.Button](g)
		button.Text = "Add " + elemType.String()
		button.Icon = theme.ContentAddIcon()
		button.OnTapped = func() { g.fieldSliceAdd(field, nil) }
	}
}
