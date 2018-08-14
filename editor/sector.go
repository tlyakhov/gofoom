package main

import (
	"math"
	"reflect"

	"github.com/gotk3/gotk3/cairo"
	"github.com/tlyakhov/gofoom/behaviors"
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/entities"
)

func DrawEntityAngle(cr *cairo.Context, e *core.PhysicalEntity) {
	cr.SetLineWidth(2)
	cr.NewPath()
	cr.MoveTo(e.Pos.X, e.Pos.Y)
	cr.LineTo(e.Pos.X+math.Cos(e.Angle*math.Pi/180.0)*e.BoundingRadius*2,
		e.Pos.Y+math.Sin(e.Angle*math.Pi/180.0)*e.BoundingRadius*2)
	cr.ClosePath()
	cr.Stroke()
}
func DrawEntity(cr *cairo.Context, e core.AbstractEntity) {
	phys := e.Physical()

	if _, ok := e.(*entities.Player); ok {
		// Let's get fancy:
		cr.SetSourceRGB(0.6, 0.6, 0.6)
		cr.NewPath()
		cr.Arc(phys.Pos.X, phys.Pos.Y, phys.BoundingRadius/2, 0, math.Pi*2)
		cr.ClosePath()
		cr.Stroke()
		cr.SetSourceRGB(0.33, 0.33, 0.33)
		DrawEntityAngle(cr, phys)
	} else if _, ok := e.(*entities.Light); ok {
		for _, b := range e.Behaviors() {
			if lb, ok := b.(*behaviors.Light); ok {
				cr.SetSourceRGB(lb.Diffuse.X, lb.Diffuse.Y, lb.Diffuse.Z)
			}
		}
	} // Sprite...

	hovering := editor.HoveringObjects[e.GetBase().ID] == e
	selected := editor.SelectedObjects[e.GetBase().ID] == e
	if selected {
		cr.SetSourceRGB(ColorSelectionPrimary.X, ColorSelectionPrimary.Y, ColorSelectionPrimary.Z)
	} else if hovering {
		cr.SetSourceRGB(ColorSelectionSecondary.X, ColorSelectionSecondary.Y, ColorSelectionSecondary.Z)
	}

	cr.NewPath()
	cr.Arc(phys.Pos.X, phys.Pos.Y, phys.BoundingRadius, 0, math.Pi*2)
	cr.ClosePath()
	cr.Stroke()
}

func DrawSector(cr *cairo.Context, sector core.AbstractSector) {
	phys := sector.Physical()

	if len(phys.Segments) == 0 {
		return
	}

	cr.Save()

	sectorHovering := editor.HoveringObjects[sector.GetBase().ID] == sector
	sectorSelected := editor.SelectedObjects[sector.GetBase().ID] == sector

	if editor.EntitiesVisible {
		for _, e := range phys.Entities {
			DrawEntity(cr, e)
		}
	}

	for i, segment := range phys.Segments {
		next := phys.Segments[(i+1)%len(phys.Segments)]

		segmentHovering := editor.HoveringObjects[segment.GetBase().ID] == segment
		segmentSelected := editor.SelectedObjects[segment.GetBase().ID] == segment

		if segment.AdjacentSector == nil {
			cr.SetSourceRGB(1, 1, 1)
		} else {
			cr.SetSourceRGB(1, 1, 0)
		}

		if sectorHovering || sectorSelected {
			if segment.AdjacentSector == nil {
				cr.SetSourceRGB(ColorSelectionPrimary.X, ColorSelectionPrimary.Y, ColorSelectionPrimary.Z)
			} else {
				cr.SetSourceRGB(ColorSelectionSecondary.X, ColorSelectionSecondary.Y, ColorSelectionSecondary.Z)
			}
		} else if segmentHovering {
			cr.SetSourceRGB(ColorSelectionSecondary.X, ColorSelectionSecondary.Y, ColorSelectionSecondary.Z)
		} else if segmentSelected {
			cr.SetSourceRGB(ColorSelectionPrimary.X, ColorSelectionPrimary.Y, ColorSelectionPrimary.Z)
		}

		// Highlight PVS sectors...
		for _, obj := range editor.SelectedObjects {
			s2, ok := obj.(core.AbstractSector)
			if !ok || s2 == sector {
				continue
			}
			if s2.Physical().PVS[sector.GetBase().ID] != nil {
				cr.SetSourceRGB(ColorPVS.X, ColorPVS.Y, ColorPVS.Z)
			}
		}

		// Draw segment
		cr.SetLineWidth(1)
		cr.NewPath()
		cr.MoveTo(segment.A.X, segment.A.Y)
		cr.LineTo(next.A.X, next.A.Y)
		cr.ClosePath()
		cr.Stroke()
		// Draw normal
		cr.NewPath()
		ns := next.A.Add(segment.A).Mul(0.5)
		ne := ns.Add(segment.Normal.Mul(4.0))
		cr.MoveTo(ns.X, ns.Y)
		cr.LineTo(ne.X, ne.Y)
		cr.ClosePath()
		cr.Stroke()

		_, mapPointHovering := editor.HoveringObjects[segment.ID].(*MapPoint)
		_, mapPointSelected := editor.SelectedObjects[segment.ID].(*MapPoint)

		if mapPointSelected {
			cr.SetSourceRGB(ColorSelectionPrimary.X, ColorSelectionPrimary.Y, ColorSelectionPrimary.Z)
			DrawHandle(cr, segment.A)
		} else if mapPointHovering {
			cr.SetSourceRGB(ColorSelectionSecondary.X, ColorSelectionSecondary.Y, ColorSelectionSecondary.Z)
			DrawHandle(cr, segment.A)
		} else {
			cr.Rectangle(segment.A.X-1, segment.A.Y-1, 2, 2)
			cr.Stroke()
		}
	}

	if editor.SectorTypesVisible {
		text := reflect.TypeOf(sector).String()
		extents := cr.TextExtents(text)
		cr.Save()
		cr.SetSourceRGB(0.3, 0.3, 0.3)
		cr.Translate(phys.Center.X-extents.Width/2, phys.Center.Y-extents.Height/2+extents.YBearing)
		cr.ShowText(text)
		cr.Restore()
	}

	cr.Restore()
}
