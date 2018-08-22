package main

import (
	"reflect"

	"github.com/gotk3/gotk3/cairo"
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/core"
)

func DrawHandle(cr *cairo.Context, v concepts.Vector2) {
	v = editor.WorldToScreen(v)
	v1 := editor.ScreenToWorld(v.Sub(concepts.Vector2{3, 3}))
	v2 := editor.ScreenToWorld(v.Add(concepts.Vector2{3, 3}))
	cr.Rectangle(v1.X, v1.Y, v2.X-v1.X, v2.Y-v1.Y)
	cr.Stroke()
}

func DrawSector(cr *cairo.Context, sector core.AbstractSector) {
	phys := sector.Physical()

	if len(phys.Segments) == 0 {
		return
	}

	cr.Save()

	sectorHovering := indexOfObject(editor.HoveringObjects, sector) != -1
	sectorSelected := indexOfObject(editor.SelectedObjects, sector) != -1

	if editor.EntitiesVisible {
		for _, e := range phys.Entities {
			DrawEntity(cr, e)
		}
	}

	for i, segment := range phys.Segments {
		next := phys.Segments[(i+1)%len(phys.Segments)]

		if next.A == segment.A {
			continue
		}

		segmentHovering := indexOfObject(editor.HoveringObjects, segment) != -1
		segmentSelected := indexOfObject(editor.SelectedObjects, segment) != -1

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

		mapPointHovering := indexOfObject(editor.HoveringObjects, &MapPoint{Segment: segment}) != -1
		mapPointSelected := indexOfObject(editor.SelectedObjects, &MapPoint{Segment: segment}) != -1

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
