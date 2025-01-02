// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"log"
	"reflect"
	"strconv"

	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2/widget"
	"github.com/spf13/cast"
)

func (g *Grid) fieldNormal(field *state.PropertyGridField) {
	origCartesian := ""
	origPhi := ""
	origTheta := ""
	for i, v := range field.Values {
		if i != 0 {
			origCartesian += ", "
			origPhi += ", "
			origTheta += ", "
		}

		cartesian := v.Value.Interface().(*concepts.Vector3)
		_, theta, phi := cartesian.ToSpherical()
		origCartesian += cartesian.StringHuman(4)
		origPhi += strconv.FormatFloat(phi*concepts.Rad2deg, 'G', 4, 64)
		origTheta += strconv.FormatFloat(theta*concepts.Rad2deg, 'G', 4, 64)
	}

	entryCartesian := widget.NewEntry()
	entryCartesian.SetText(origCartesian)
	entryCartesian.OnSubmitted = func(text string) {
		parsed, err := concepts.ParseVector3(text)
		if err != nil {
			log.Printf("Couldn't parse Vector3 from user entry. %v\n", err)
			entryCartesian.SetText(origCartesian)
			g.Focus(g.GridWidget)
			return
		}

		g.ApplySetPropertyAction(field, reflect.ValueOf(parsed).Elem())
		origCartesian = parsed.StringHuman(4)
		g.Focus(g.GridWidget)
	}

	entryPhi := widget.NewEntry()
	entryPhi.SetText(origPhi)
	entryPhi.OnSubmitted = func(text string) {
		parsed, err := cast.ToFloat64E(text)
		if err != nil {
			log.Printf("Couldn't parse number: %v\n", err)
			entryPhi.SetText(origPhi)
			g.Focus(g.GridWidget)
			return
		}

		var cartesian *concepts.Vector3
		for _, v := range field.Values {
			cartesian = v.Value.Interface().(*concepts.Vector3)
			_, theta, _ := cartesian.ToSpherical()
			cartesian.FromSpherical(1.0, theta, parsed*concepts.Deg2rad)
		}

		g.ApplySetPropertyAction(field, reflect.ValueOf(cartesian).Elem())
		g.Focus(g.GridWidget)
	}

	entrySlope := widget.NewEntry()
	entrySlope.SetText(origTheta)
	entrySlope.OnSubmitted = func(text string) {
		parsed, err := cast.ToFloat64E(text)
		if err != nil {
			log.Printf("Couldn't parse number: %v\n", err)
			entrySlope.SetText(origTheta)
			g.Focus(g.GridWidget)
			return
		}

		var cartesian *concepts.Vector3
		for _, v := range field.Values {
			cartesian = v.Value.Interface().(*concepts.Vector3)
			_, _, phi := cartesian.ToSpherical()
			cartesian.FromSpherical(1.0, parsed*concepts.Deg2rad, phi)
		}

		g.ApplySetPropertyAction(field, reflect.ValueOf(cartesian).Elem())
		g.Focus(g.GridWidget)
	}

	f := gridAddOrUpdateWidgetAtIndex[*widget.Form](g)
	f.Items = make([]*widget.FormItem, 0)
	f.Append("Cartesian", entryCartesian)
	f.Append("Azimuth (°)", entryPhi)
	f.Append("Polar (°)", entrySlope)
	f.Refresh()

}
