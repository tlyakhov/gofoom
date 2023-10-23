package main

import (
	"math"
	"reflect"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/components/behaviors"

	"github.com/gotk3/gotk3/cairo"
)

func DrawMobAngle(cr *cairo.Context, e *core.Mob) {
	astart := (e.Angle - editor.Renderer.FOV/2) * math.Pi / 180.0
	aend := (e.Angle + editor.Renderer.FOV/2) * math.Pi / 180.0
	cr.SetLineWidth(2)
	cr.NewPath()
	cr.MoveTo(e.Pos.Now[0], e.Pos.Now[1])
	cr.LineTo(e.Pos.Now[0]+math.Cos(astart)*e.BoundingRadius*2, e.Pos.Now[1]+math.Sin(astart)*e.BoundingRadius*2)
	cr.Arc(e.Pos.Now[0], e.Pos.Now[1], e.BoundingRadius*2, astart, aend)
	cr.MoveTo(e.Pos.Now[0], e.Pos.Now[1])
	cr.LineTo(e.Pos.Now[0]+math.Cos(aend)*e.BoundingRadius*2, e.Pos.Now[1]+math.Sin(aend)*e.BoundingRadius*2)
	cr.ClosePath()
	cr.Stroke()
}

func DrawMob(cr *cairo.Context, mober *concepts.EntityRef) {
	mob := core.MobFromDb(mober)

	cr.SetSourceRGB(0.6, 0.6, 0.6)
	if player := behaviors.PlayerFromDb(mober); player != nil {
		// Let's get fancy:
		cr.SetSourceRGB(0.6, 0.6, 0.6)
		cr.SetLineWidth(1)
		cr.NewPath()
		cr.Arc(mob.Pos.Now[0], mob.Pos.Now[1], mob.BoundingRadius/2, 0, math.Pi*2)
		cr.ClosePath()
		cr.Stroke()
		cr.SetSourceRGB(0.33, 0.33, 0.33)
		DrawMobAngle(cr, mob)
	} else if light := core.LightFromDb(mober); light != nil {
		cr.SetSourceRGB(light.Diffuse[0], light.Diffuse[1], light.Diffuse[2])
	} // Sprite...

	hovering := state.IndexOf(editor.HoveringObjects, mob) != -1
	selected := state.IndexOf(editor.SelectedObjects, mob) != -1
	if selected {
		cr.SetSourceRGB(ColorSelectionPrimary[0], ColorSelectionPrimary[1], ColorSelectionPrimary[2])
	} else if hovering {
		cr.SetSourceRGB(ColorSelectionSecondary[0], ColorSelectionSecondary[1], ColorSelectionSecondary[2])
	}

	cr.SetLineWidth(1)
	cr.NewPath()
	cr.Arc(mob.Pos.Now[0], mob.Pos.Now[1], mob.BoundingRadius, 0, math.Pi*2)
	cr.ClosePath()
	cr.Stroke()

	if editor.MobTypesVisible {
		text := reflect.TypeOf(mob).String()
		extents := cr.TextExtents(text)
		cr.Save()
		cr.SetSourceRGB(0.3, 0.3, 0.3)
		cr.Translate(mob.Pos.Now[0]-extents.Width/2, mob.Pos.Now[1]-extents.Height/2-extents.YBearing)
		cr.ShowText(text)
		cr.Restore()
	}
}
