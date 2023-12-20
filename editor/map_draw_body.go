package main

import (
	"math"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/components/behaviors"

	"github.com/gotk3/gotk3/cairo"
)

func DrawBodyAngle(cr *cairo.Context, e *core.Body) {
	astart := (e.Angle.Now - editor.Renderer.FOV/2) * math.Pi / 180.0
	aend := (e.Angle.Now + editor.Renderer.FOV/2) * math.Pi / 180.0
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

func DrawBody(cr *cairo.Context, ibody *concepts.EntityRef) {
	body := core.BodyFromDb(ibody)

	if body == nil {
		return
	}

	cr.SetSourceRGB(0.6, 0.6, 0.6)
	if player := behaviors.PlayerFromDb(ibody); player != nil {
		// Let's get fancy:
		cr.SetSourceRGB(0.6, 0.6, 0.6)
		cr.SetLineWidth(1)
		cr.NewPath()
		cr.Arc(body.Pos.Now[0], body.Pos.Now[1], body.BoundingRadius/2, 0, math.Pi*2)
		cr.ClosePath()
		cr.Stroke()
	} else if light := core.LightFromDb(ibody); light != nil {
		cr.SetSourceRGB(light.Diffuse[0], light.Diffuse[1], light.Diffuse[2])
	} // Sprite...

	hovering := state.IndexOf(editor.HoveringObjects, ibody) != -1
	selected := state.IndexOf(editor.SelectedObjects, ibody) != -1
	if selected {
		cr.SetSourceRGB(ColorSelectionPrimary[0], ColorSelectionPrimary[1], ColorSelectionPrimary[2])
	} else if hovering {
		cr.SetSourceRGB(ColorSelectionSecondary[0], ColorSelectionSecondary[1], ColorSelectionSecondary[2])
	}

	cr.SetLineWidth(1)
	cr.NewPath()
	cr.Arc(body.Pos.Now[0], body.Pos.Now[1], body.BoundingRadius, 0, math.Pi*2)
	cr.ClosePath()
	cr.Stroke()
	cr.SetSourceRGB(0.33, 0.33, 0.33)
	DrawBodyAngle(cr, body)

	if editor.ComponentNamesVisible {
		text := ibody.NameString()
		extents := cr.TextExtents(text)
		cr.Save()
		cr.SetSourceRGB(0.3, 0.3, 0.5)
		cr.Translate(body.Pos.Now[0]-extents.Width/2, body.Pos.Now[1]-extents.Height/2-extents.YBearing)
		cr.ShowText(text)
		cr.Restore()
	}
}
