// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"log"
	"reflect"
	"strconv"
	"strings"

	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/spf13/cast"
)

func (g *Grid) fieldNormal(field *state.PropertyGridField) {
	origCartesian := ""
	var origPhi strings.Builder
	var origTheta strings.Builder
	for i, v := range field.Values {
		if i != 0 {
			origCartesian += ", "
			origPhi.WriteString(", ")
			origTheta.WriteString(", ")
		}

		cartesian := v.Value.Interface().(*concepts.Vector3)
		_, theta, phi := cartesian.ToSpherical()
		origCartesian += cartesian.StringHuman(4)
		origPhi.WriteString(strconv.FormatFloat(phi*concepts.Rad2deg, 'G', 4, 64))
		origTheta.WriteString(strconv.FormatFloat(theta*concepts.Rad2deg, 'G', 4, 64))
	}

	entryCartesian := widget.NewEntry()
	entryCartesian.SetText(origCartesian)
	log.Printf("entryCartesian: %v", origCartesian)
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
	entryPhi.SetText(origPhi.String())
	entryPhi.OnSubmitted = func(text string) {
		parsed, err := cast.ToFloat64E(text)
		if err != nil {
			log.Printf("Couldn't parse number: %v\n", err)
			entryPhi.SetText(origPhi.String())
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
	entrySlope.SetText(origTheta.String())
	entrySlope.OnSubmitted = func(text string) {
		parsed, err := cast.ToFloat64E(text)
		if err != nil {
			log.Printf("Couldn't parse number: %v\n", err)
			entrySlope.SetText(origTheta.String())
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

	if field.Disabled() {
		entryCartesian.Disable()
		entryPhi.Disable()
		entrySlope.Disable()
	} else {
		entryCartesian.Enable()
		entryPhi.Enable()
		entrySlope.Enable()
	}

	f := gridAddOrUpdateWidgetAtIndex[*widget.Form](g)
	fyne.Do(func() {
		f.Append("Cartesian", entryCartesian)
		f.Append("Azimuth (°)", entryPhi)
		f.Append("Polar (°)", entrySlope)
	})
}
