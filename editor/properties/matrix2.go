// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"reflect"
	"strconv"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *Grid) fieldMatrix2Aspect(field *state.PropertyGridField, scaleHeight bool) {
	action := &actions.SetProperty{
		IEditor:           g.IEditor,
		PropertyGridField: field,
		ValuesToAssign:    make([]reflect.Value, len(field.Values)),
	}
	// For Identity matrix transforms, the texture will fill the dimensions of
	// the surface it's on. This method calculates the aspect ratio of the
	// surface in the world, and adjusts the transform accordingly.
	for i, v := range field.Values {
		var worldWidth, worldHeight float64
		grandparent := v.Ancestors[len(v.Ancestors)-2]
		switch typed := grandparent.(type) {
		case *core.Sector:
			// This is a floor or ceiling
			worldWidth = typed.Max[0] - typed.Min[0]
			worldHeight = typed.Max[1] - typed.Min[1]
		case *core.SectorSegment:
			fz, cz := typed.Sector.SlopedZNow(typed.A)
			worldWidth = typed.Length
			worldHeight = cz - fz
		case *core.InternalSegment:
			worldWidth = typed.Length
			worldHeight = typed.Top - typed.Bottom
		}
		if worldHeight == 0 || worldWidth == 0 {
			continue
		}
		newTransform := &concepts.Matrix2{1, 0, 0, 1, 0, 0}
		if scaleHeight {
			newTransform[3] = worldHeight / worldWidth
		} else {
			newTransform[0] = worldWidth / worldHeight
		}
		action.ValuesToAssign[i] = reflect.ValueOf(newTransform).Elem()
	}
	g.NewAction(action)
	action.Act()
}

func (g *Grid) fieldMatrix2(field *state.PropertyGridField) {
	origMatrix := ""
	origDelta := ""
	origAngle := ""
	origScale := ""
	for i, v := range field.Values {
		if i != 0 {
			origMatrix += ", "
			origDelta += ", "
			origAngle += ", "
			origScale += ", "
		}
		m := v.Value.Interface().(*concepts.Matrix2)
		origMatrix += m.String()
		a, t, s := m.GetTransform()
		origDelta += t.String()
		origAngle += strconv.FormatFloat(a, 'f', 2, 64)
		origScale += s.String()
	}

	f := gridAddOrUpdateWidgetAtIndex[*widget.Form](g)
	f.Items = make([]*widget.FormItem, 0)
	entry := widget.NewEntry()
	entry.SetText(origMatrix)
	entry.OnSubmitted = func(text string) {
		if toSet, err := concepts.ParseMatrix2(text); err == nil {
			g.ApplySetPropertyAction(field, reflect.ValueOf(toSet).Elem())
			origMatrix = text
		} else {
			entry.SetText(origMatrix)
		}
	}
	f.Append("Matrix", entry)
	eDelta := widget.NewEntry()
	eDelta.OnSubmitted = func(text string) {
		var delta *concepts.Vector2
		var err error
		if delta, err = concepts.ParseVector2(text); err != nil {
			eDelta.SetText(origDelta)
			return
		}
		action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ValuesToAssign: make([]reflect.Value, len(field.Values))}
		for i, v := range field.Values {
			m := v.Value.Interface().(*concepts.Matrix2)
			m = m.Translate(delta)
			action.ValuesToAssign[i] = reflect.ValueOf(m).Elem()
		}
		g.NewAction(action)
		action.Act()
	}
	eDelta.SetText(origDelta)
	f.Append("DX/DY", eDelta)
	eAngle := widget.NewEntry()
	eAngle.SetText(origAngle)
	f.Append("Angle", eAngle)
	eScale := widget.NewEntry()
	eScale.SetText(origScale)
	f.Append("SX/SY", eScale)
	eScale.OnSubmitted = func(text string) {
		var scale *concepts.Vector2
		var err error
		if scale, err = concepts.ParseVector2(text); err != nil {
			eScale.SetText(origScale)
			return
		}
		action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ValuesToAssign: make([]reflect.Value, len(field.Values))}
		for i, v := range field.Values {
			m := v.Value.Interface().(*concepts.Matrix2)
			m = m.AxisScale(scale)
			action.ValuesToAssign[i] = reflect.ValueOf(m).Elem()
		}
		g.NewAction(action)
		action.Act()
	}

	switch field.Values[0].Parent().(type) {
	case *materials.Surface:
		f.Append("", widget.NewButtonWithIcon("Reset/Fill", theme.ContentClearIcon(), func() {
			toSet := &concepts.Matrix2{}
			toSet.SetIdentity()
			g.ApplySetPropertyAction(field, reflect.ValueOf(toSet).Elem())
		}))
		f.Append("Aspect", container.NewHBox(
			widget.NewButtonWithIcon("Scale W", theme.ViewFullScreenIcon(), func() { g.fieldMatrix2Aspect(field, false) }),
			widget.NewButtonWithIcon("Scale H", theme.ViewFullScreenIcon(), func() { g.fieldMatrix2Aspect(field, true) })))
	}
}
