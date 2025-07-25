// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"math"
	"reflect"
	"strconv"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *Grid) fieldMatrix2Aspect(field *state.PropertyGridField, scaleHeight bool) {
	action := &actions.SetProperty{
		Action:            state.Action{IEditor: g},
		PropertyGridField: field,
		ValuesToAssign:    make([]reflect.Value, len(field.Values)),
	}
	// For Identity matrix transforms, the texture will fill the dimensions of
	// the surface it's on. This method calculates the aspect ratio of the
	// surface in the world, and adjusts the transform accordingly.
	for i, v := range field.Values {
		var worldWidth, worldHeight float64
		grandparent := v.Ancestors[len(v.Ancestors)-3]
		switch typed := grandparent.(type) {
		case *core.Sector:
			// This is a floor or ceiling
			worldWidth = typed.Max[0] - typed.Min[0]
			worldHeight = typed.Max[1] - typed.Min[1]
		case *core.SectorSegment:
			fz, cz := typed.Sector.ZAt(dynamic.DynamicNow, typed.A)
			worldWidth = typed.Length
			worldHeight = cz - fz
			switch field.Name {
			case "Segment.Low.ℝ²→ℝ².Spawn":
				if typed.AdjacentSegment != nil {
					adj := core.GetSector(typed.AdjacentSector)
					afz := adj.Bottom.ZAt(dynamic.DynamicNow, typed.A)
					worldHeight = math.Abs(fz - afz)
				}
			case "Segment.High.ℝ²→ℝ².Spawn":
				if typed.AdjacentSegment != nil {
					adj := core.GetSector(typed.AdjacentSector)
					acz := adj.Top.ZAt(dynamic.DynamicNow, typed.A)
					worldHeight = math.Abs(cz - acz)
				}
			}
		case *core.InternalSegment:
			worldWidth = typed.Length
			worldHeight = typed.Top - typed.Bottom
		}
		if worldHeight == 0 || worldWidth == 0 {
			action.ValuesToAssign[i] = v.Deref()
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
	g.Act(action)
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
	eDelta := widget.NewEntry()
	eDelta.OnSubmitted = func(text string) {
		var delta *concepts.Vector2
		var err error
		if delta, err = concepts.ParseVector2(text); err != nil {
			eDelta.SetText(origDelta)
			return
		}
		action := &actions.SetProperty{
			Action:            state.Action{IEditor: g},
			PropertyGridField: field,
			ValuesToAssign:    make([]reflect.Value, len(field.Values)),
		}
		for i, v := range field.Values {
			m := v.Value.Interface().(*concepts.Matrix2)
			m = m.TranslateSelf(delta)
			action.ValuesToAssign[i] = reflect.ValueOf(m).Elem()
		}
		g.Act(action)
	}
	eDelta.SetText(origDelta)
	eAngle := widget.NewEntry()
	eAngle.SetText(origAngle)
	eScale := widget.NewEntry()
	eScale.SetText(origScale)
	eScale.OnSubmitted = func(text string) {
		var scale *concepts.Vector2
		var err error
		if scale, err = concepts.ParseVector2(text); err != nil {
			eScale.SetText(origScale)
			return
		}
		action := &actions.SetProperty{
			Action:            state.Action{IEditor: g},
			PropertyGridField: field,
			ValuesToAssign:    make([]reflect.Value, len(field.Values)),
		}
		for i, v := range field.Values {
			m := v.Value.Interface().(*concepts.Matrix2)
			m = m.AxisScale(scale)
			action.ValuesToAssign[i] = reflect.ValueOf(m).Elem()
		}
		g.Act(action)
	}

	bReset := widget.NewButtonWithIcon("Reset/Fill", theme.ContentClearIcon(), func() {
		toSet := &concepts.Matrix2{}
		toSet.SetIdentity()
		g.ApplySetPropertyAction(field, reflect.ValueOf(toSet).Elem())
	})
	bScaleW := widget.NewButtonWithIcon("Scale W", theme.ViewFullScreenIcon(), func() { g.fieldMatrix2Aspect(field, false) })
	bScaleH := widget.NewButtonWithIcon("Scale H", theme.ViewFullScreenIcon(), func() { g.fieldMatrix2Aspect(field, true) })

	if field.Disabled() {
		eDelta.Disable()
		eAngle.Disable()
		eScale.Disable()
		bReset.Disable()
		bScaleW.Disable()
		bScaleH.Disable()
	} else {
		eDelta.Enable()
		eAngle.Enable()
		eScale.Enable()
		bReset.Enable()
		bScaleW.Enable()
		bScaleH.Enable()
	}
	f := gridAddOrUpdateWidgetAtIndex[*widget.Form](g)
	fyne.Do(func() {
		f.Items = make([]*widget.FormItem, 0)
		f.Append("Matrix", entry)
		f.Append("DX/DY", eDelta)
		f.Append("Angle", eAngle)
		f.Append("SX/SY", eScale)
		ancestors := field.Values[0].Ancestors
		if len(ancestors) >= 2 {
			switch ancestors[len(ancestors)-2].(type) {
			case *materials.Surface:
				f.Append("", bReset)
				f.Append("Aspect", container.NewHBox(bScaleW, bScaleH))
			}
		}
		f.Refresh()
	})
}
