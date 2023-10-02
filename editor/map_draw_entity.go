package main

import (
	"math"
	"reflect"

	"tlyakhov/gofoom/concepts"

	"tlyakhov/gofoom/behaviors"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/entities"

	"github.com/gotk3/gotk3/cairo"
)

func DrawEntityAngle(cr *cairo.Context, e *core.PhysicalEntity) {
	astart := (e.Angle - editor.Renderer.FOV/2) * math.Pi / 180.0
	aend := (e.Angle + editor.Renderer.FOV/2) * math.Pi / 180.0
	cr.SetLineWidth(2)
	cr.NewPath()
	cr.MoveTo(e.Pos.X, e.Pos.Y)
	cr.LineTo(e.Pos.X+math.Cos(astart)*e.BoundingRadius*2, e.Pos.Y+math.Sin(astart)*e.BoundingRadius*2)
	cr.Arc(e.Pos.X, e.Pos.Y, e.BoundingRadius*2, astart, aend)
	cr.MoveTo(e.Pos.X, e.Pos.Y)
	cr.LineTo(e.Pos.X+math.Cos(aend)*e.BoundingRadius*2, e.Pos.Y+math.Sin(aend)*e.BoundingRadius*2)
	cr.ClosePath()
	cr.Stroke()
}

func DrawEntity(cr *cairo.Context, e core.AbstractEntity) {
	phys := e.Physical()

	cr.SetSourceRGB(0.6, 0.6, 0.6)
	if _, ok := e.(*entities.Player); ok {
		// Let's get fancy:
		cr.SetSourceRGB(0.6, 0.6, 0.6)
		cr.SetLineWidth(1)
		cr.NewPath()
		cr.Arc(phys.Pos.X, phys.Pos.Y, phys.BoundingRadius/2, 0, math.Pi*2)
		cr.ClosePath()
		cr.Stroke()
		cr.SetSourceRGB(0.33, 0.33, 0.33)
		DrawEntityAngle(cr, phys)
	} else if _, ok := e.(*entities.Light); ok {
		for _, b := range e.Physical().Behaviors {
			if lb, ok := b.(*behaviors.Light); ok {
				cr.SetSourceRGB(lb.Diffuse.X, lb.Diffuse.Y, lb.Diffuse.Z)
			}
		}
	} // Sprite...

	hovering := concepts.IndexOf(editor.HoveringObjects, e) != -1
	selected := concepts.IndexOf(editor.SelectedObjects, e) != -1
	if selected {
		cr.SetSourceRGB(ColorSelectionPrimary.X, ColorSelectionPrimary.Y, ColorSelectionPrimary.Z)
	} else if hovering {
		cr.SetSourceRGB(ColorSelectionSecondary.X, ColorSelectionSecondary.Y, ColorSelectionSecondary.Z)
	}

	cr.SetLineWidth(1)
	cr.NewPath()
	cr.Arc(phys.Pos.X, phys.Pos.Y, phys.BoundingRadius, 0, math.Pi*2)
	cr.ClosePath()
	cr.Stroke()

	if editor.EntityTypesVisible {
		text := reflect.TypeOf(e).String()
		extents := cr.TextExtents(text)
		cr.Save()
		cr.SetSourceRGB(0.3, 0.3, 0.3)
		cr.Translate(phys.Pos.X-extents.Width/2, phys.Pos.Y-extents.Height/2-extents.YBearing)
		cr.ShowText(text)
		cr.Restore()
	}
}
