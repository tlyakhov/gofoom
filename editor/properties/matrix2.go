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
	// For Identity matrix transforms, the texture will fill the dimensions of
	// the surface it's on. This method calculates the aspect ratio of the
	// surface in the world, and adjusts the transform accordingly.
	for _, v := range field.Values {
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
		action := &actions.SetProperty{
			IEditor:           g.IEditor,
			PropertyGridField: field,
			ToSet:             reflect.ValueOf(newTransform).Elem(),
		}
		g.NewAction(action)
		action.Act()
	}
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
			action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: reflect.ValueOf(toSet).Elem()}
			g.NewAction(action)
			action.Act()
			origMatrix = text
		} else {
			entry.SetText(origMatrix)
		}
	}
	f.Append("Matrix", entry)
	eDelta := widget.NewEntry()
	eDelta.OnSubmitted = func(text string) {
		/*	if delta, err := concepts.ParseVector2(text); err == nil {
			for _, v := range field.Values {
				m := v.Interface().(*concepts.Matrix2)
				m = m.Translate(delta)
			}
			action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: reflect.ValueOf(toSet).Elem()}
			g.NewAction(action)
			action.Act()
		}*/
	}
	eDelta.SetText(origDelta)
	f.Append("DX/DY", eDelta)
	eAngle := widget.NewEntry()
	eAngle.SetText(origAngle)
	f.Append("Angle", eAngle)
	eScale := widget.NewEntry()
	eScale.SetText(origScale)
	f.Append("SX/SY", eScale)

	switch field.Values[0].Parent().(type) {
	case *materials.Surface:
		f.Append("", widget.NewButtonWithIcon("Reset/Fill", theme.ContentClearIcon(), func() {
			toSet := &concepts.Matrix2{}
			toSet.SetIdentity()
			action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: reflect.ValueOf(toSet).Elem()}
			g.NewAction(action)
			action.Act()
		}))
		f.Append("Aspect", container.NewHBox(
			widget.NewButtonWithIcon("Scale W", theme.ViewFullScreenIcon(), func() { g.fieldMatrix2Aspect(field, false) }),
			widget.NewButtonWithIcon("Scale H", theme.ViewFullScreenIcon(), func() { g.fieldMatrix2Aspect(field, true) })))
	}
}
