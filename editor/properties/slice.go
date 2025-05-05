// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"reflect"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *Grid) fieldSliceAdd(field *state.PropertyGridField, concreteType reflect.Type) {
	g.Act(&actions.ChangeSlice{
		Action:       state.Action{IEditor: g},
		SlicePtr:     field.Values[0].Value,
		Parent:       field.Values[0].Parent,
		ConcreteType: concreteType,
		Mode:         actions.AddSliceElementMode,
		Index:        -1})
	g.Focus(g.GridWidget)
}

var animationTypes = map[string]reflect.Type{
	"Animation[int]":     reflect.TypeFor[dynamic.Animation[int]](),
	"Animation[float64]": reflect.TypeFor[dynamic.Animation[float64]](),
	"Animation[Vector2]": reflect.TypeFor[dynamic.Animation[concepts.Vector2]](),
	"Animation[Vector3]": reflect.TypeFor[dynamic.Animation[concepts.Vector3]](),
	"Animation[Vector4]": reflect.TypeFor[dynamic.Animation[concepts.Vector4]](),
}

func (g *Grid) fieldSlice(field *state.PropertyGridField) {
	// field.Type is *[]<something>
	elemType := field.Type.Elem().Elem()
	if field.Type.Elem().Kind() == reflect.Array {
		label := gridAddOrUpdateWidgetAtIndex[*widget.Label](g)
		label.Text = "Fixed Array"
		return
	}
	if elemType == reflect.TypeFor[dynamic.Animated]() {
		buttons := make([]fyne.CanvasObject, len(animationTypes))
		i := 0
		for name, t := range animationTypes {
			t := t // To ensure correct scope for closure
			b := widget.NewButtonWithIcon("Add "+name, theme.ContentAddIcon(), func() { g.fieldSliceAdd(field, t) })
			if field.Disabled() {
				b.Disable()
			} else {
				b.Enable()
			}
			buttons[i] = b
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
		if field.Disabled() {
			button.Disable()
		} else {
			button.Enable()
		}
	}
}

func (g *Grid) fieldChangeSlice(field *state.PropertyGridField) {
	// field.Type is *[]<something>
	//elemType := field.Type.Elem().Elem()
	buttonDec := widget.NewButtonWithIcon("Up", theme.MoveUpIcon(), func() {
		g.Act(&actions.ChangeSlice{
			Action:   state.Action{IEditor: g},
			SlicePtr: field.Values[0].Value,
			Parent:   field.Values[0].Parent,
			Mode:     actions.DecSliceElementMode,
			Index:    field.SliceIndex})
		g.Focus(g.GridWidget)
	})
	buttonInc := widget.NewButtonWithIcon("Down", theme.MoveDownIcon(), func() {
		g.Act(&actions.ChangeSlice{
			Action:   state.Action{IEditor: g},
			SlicePtr: field.Values[0].Value,
			Parent:   field.Values[0].Parent,
			Mode:     actions.IncSliceElementMode,
			Index:    field.SliceIndex})
		g.Focus(g.GridWidget)
	})
	buttonDelete := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		g.Act(&actions.ChangeSlice{
			Action:   state.Action{IEditor: g},
			SlicePtr: field.Values[0].Value,
			Parent:   field.Values[0].Parent,
			Mode:     actions.DeleteSliceElementMode,
			Index:    field.SliceIndex})
		g.Focus(g.GridWidget)
	})

	if field.Disabled() {
		buttonDec.Disable()
		buttonInc.Disable()
		buttonDelete.Disable()
	} else {
		buttonDec.Enable()
		buttonInc.Enable()
		buttonDelete.Enable()
	}
	c := gridAddOrUpdateWidgetAtIndex[*fyne.Container](g)
	c.Layout = layout.NewHBoxLayout()
	c.Objects = []fyne.CanvasObject{}
	if field.SliceIndex != 0 {
		c.Objects = append(c.Objects, buttonDec)
	}
	if field.SliceIndex < field.Values[0].Value.Elem().Len()-1 {
		c.Objects = append(c.Objects, buttonInc)
	}
	if field.Type.Kind() == reflect.Slice {
		c.Objects = append(c.Objects, buttonDelete)
	}
}
