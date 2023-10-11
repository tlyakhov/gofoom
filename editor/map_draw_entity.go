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
	cr.MoveTo(e.Pos[0], e.Pos[1])
	cr.LineTo(e.Pos[0]+math.Cos(astart)*e.BoundingRadius*2, e.Pos[1]+math.Sin(astart)*e.BoundingRadius*2)
	cr.Arc(e.Pos[0], e.Pos[1], e.BoundingRadius*2, astart, aend)
	cr.MoveTo(e.Pos[0], e.Pos[1])
	cr.LineTo(e.Pos[0]+math.Cos(aend)*e.BoundingRadius*2, e.Pos[1]+math.Sin(aend)*e.BoundingRadius*2)
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
		cr.Arc(phys.Pos[0], phys.Pos[1], phys.BoundingRadius/2, 0, math.Pi*2)
		cr.ClosePath()
		cr.Stroke()
		cr.SetSourceRGB(0.33, 0.33, 0.33)
		DrawEntityAngle(cr, phys)
	} else if _, ok := e.(*entities.Light); ok {
		for _, b := range e.Physical().Behaviors {
			if lb, ok := b.(*behaviors.Light); ok {
				cr.SetSourceRGB(lb.Diffuse[0], lb.Diffuse[1], lb.Diffuse[2])
			}
		}
	} // Sprite...

	hovering := concepts.IndexOf(editor.HoveringObjects, e) != -1
	selected := concepts.IndexOf(editor.SelectedObjects, e) != -1
	if selected {
		cr.SetSourceRGB(ColorSelectionPrimary[0], ColorSelectionPrimary[1], ColorSelectionPrimary[2])
	} else if hovering {
		cr.SetSourceRGB(ColorSelectionSecondary[0], ColorSelectionSecondary[1], ColorSelectionSecondary[2])
	}

	cr.SetLineWidth(1)
	cr.NewPath()
	cr.Arc(phys.Pos[0], phys.Pos[1], phys.BoundingRadius, 0, math.Pi*2)
	cr.ClosePath()
	cr.Stroke()

	if editor.EntityTypesVisible {
		text := reflect.TypeOf(e).String()
		extents := cr.TextExtents(text)
		cr.Save()
		cr.SetSourceRGB(0.3, 0.3, 0.3)
		cr.Translate(phys.Pos[0]-extents.Width/2, phys.Pos[1]-extents.Height/2-extents.YBearing)
		cr.ShowText(text)
		cr.Restore()
	}
}
