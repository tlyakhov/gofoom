package main

import (
	"reflect"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/cairo"
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

	sectorHovering := concepts.IndexOf(editor.HoveringObjects, sector) != -1
	sectorSelected := concepts.IndexOf(editor.SelectedObjects, sector) != -1

	if editor.EntitiesVisible {
		for _, e := range phys.Entities {
			DrawEntity(cr, e)
		}
	}

	for _, segment := range phys.Segments {
		if segment.P == segment.Next.P {
			continue
		}

		segmentHovering := concepts.IndexOf(editor.HoveringObjects, segment) != -1
		segmentSelected := concepts.IndexOf(editor.SelectedObjects, segment) != -1

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
		cr.MoveTo(segment.P.X, segment.P.Y)
		cr.LineTo(segment.Next.P.X, segment.Next.P.Y)
		cr.ClosePath()
		cr.Stroke()
		// Draw normal
		cr.NewPath()
		ns := segment.Next.P.Add(segment.P).Mul(0.5)
		ne := ns.Add(segment.Normal.Mul(4.0))
		cr.MoveTo(ns.X, ns.Y)
		cr.LineTo(ne.X, ne.Y)
		cr.ClosePath()
		cr.Stroke()

		mapPointHovering := concepts.IndexOf(editor.HoveringObjects, &state.MapPoint{Segment: segment}) != -1
		mapPointSelected := concepts.IndexOf(editor.SelectedObjects, &state.MapPoint{Segment: segment}) != -1

		if mapPointSelected {
			cr.SetSourceRGB(ColorSelectionPrimary.X, ColorSelectionPrimary.Y, ColorSelectionPrimary.Z)
			DrawHandle(cr, segment.P)
		} else if mapPointHovering {
			cr.SetSourceRGB(ColorSelectionSecondary.X, ColorSelectionSecondary.Y, ColorSelectionSecondary.Z)
			DrawHandle(cr, segment.P)
		} else {
			cr.Rectangle(segment.P.X-1, segment.P.Y-1, 2, 2)
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
