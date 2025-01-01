// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"log"
	"math"
	"reflect"
	"strconv"

	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2/widget"
	"github.com/spf13/cast"
)

func (g *Grid) fieldNormal(field *state.PropertyGridField) {
	origV3 := ""
	origXY := ""
	origSlope := ""
	for i, v := range field.Values {
		if i != 0 {
			origV3 += ", "
			origXY += ", "
			origSlope += ", "
		}

		v3 := v.Value.Interface().(*concepts.Vector3)
		origV3 += v3.StringHuman(4)
		xy := math.Atan2(v3[1], v3[0]) * concepts.Rad2deg
		origXY += strconv.FormatFloat(xy, 'G', 4, 64)
		l2d := math.Sqrt(v3[0]*v3[0] + v3[1]*v3[1])
		slope := math.Atan2(v3[2], l2d) * concepts.Rad2deg
		origSlope += strconv.FormatFloat(slope, 'G', 4, 64)
	}

	entryV3 := widget.NewEntry()
	entryV3.SetText(origV3)
	entryV3.OnSubmitted = func(text string) {
		parsed, err := concepts.ParseVector3(text)
		if err != nil {
			log.Printf("Couldn't parse Vector3 from user entry. %v\n", err)
			entryV3.SetText(origV3)
			g.Focus(g.GridWidget)
			return
		}

		g.ApplySetPropertyAction(field, reflect.ValueOf(parsed).Elem())
		origV3 = parsed.StringHuman(4)
		g.Focus(g.GridWidget)
	}

	entryXY := widget.NewEntry()
	entryXY.SetText(origXY)
	entryXY.OnSubmitted = func(text string) {
		parsed, err := cast.ToFloat64E(text)
		if err != nil {
			log.Printf("Couldn't parse number: %v\n", err)
			entryXY.SetText(origXY)
			g.Focus(g.GridWidget)
			return
		}

		var v3 *concepts.Vector3
		for _, v := range field.Values {
			v3 = v.Value.Interface().(*concepts.Vector3)
			l3d := v3.Length()
			l2d := math.Sqrt(v3[0]*v3[0] + v3[1]*v3[1])
			slope := math.Atan2(v3[2], l2d)
			v3[0] = math.Cos(parsed*concepts.Deg2rad) * math.Cos(slope) * l3d
			v3[1] = math.Sin(parsed*concepts.Deg2rad) * math.Cos(slope) * l3d
			v3[2] = math.Sin(slope) * l3d
			v3.NormSelf()
		}

		g.ApplySetPropertyAction(field, reflect.ValueOf(v3).Elem())
		g.Focus(g.GridWidget)
	}

	entrySlope := widget.NewEntry()
	entrySlope.SetText(origSlope)
	entrySlope.OnSubmitted = func(text string) {
		parsed, err := cast.ToFloat64E(text)
		if err != nil {
			log.Printf("Couldn't parse number: %v\n", err)
			entrySlope.SetText(origSlope)
			g.Focus(g.GridWidget)
			return
		}

		var v3 *concepts.Vector3
		for _, v := range field.Values {
			v3 = v.Value.Interface().(*concepts.Vector3)
			l3d := v3.Length()
			xyAngle := math.Atan2(v3[1], v3[0])
			v3[0] = math.Cos(xyAngle) * math.Cos(parsed*concepts.Deg2rad) * l3d
			v3[1] = math.Sin(xyAngle) * math.Cos(parsed*concepts.Deg2rad) * l3d
			v3[2] = math.Sin(parsed*concepts.Deg2rad) * l3d
			v3.NormSelf()
		}

		g.ApplySetPropertyAction(field, reflect.ValueOf(v3).Elem())
		g.Focus(g.GridWidget)
	}

	f := gridAddOrUpdateWidgetAtIndex[*widget.Form](g)
	f.Items = make([]*widget.FormItem, 0)
	f.Append("V3", entryV3)
	f.Append("XY (°)", entryXY)
	f.Append("Slope (°)", entrySlope)
	f.Refresh()

}
