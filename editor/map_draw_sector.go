package main

import (
	"reflect"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"
)

func (mw *MapWidget) DrawHandle(v *concepts.Vector2) {
	v = editor.WorldToScreen(v)
	v1 := editor.ScreenToWorld(v.Sub(&concepts.Vector2{3, 3}))
	v2 := editor.ScreenToWorld(v.Add(&concepts.Vector2{3, 3}))
	mw.Context.DrawRectangle(v1[0], v1[1], v2[0]-v1[0], v2[1]-v1[1])
	mw.Context.Stroke()
}

func (mw *MapWidget) DrawSector(sector *core.Sector) {
	if len(sector.Segments) == 0 {
		return
	}

	mw.Context.Push()

	sectorHovering := state.IndexOf(editor.HoveringObjects, sector.EntityRef) != -1
	sectorSelected := state.IndexOf(editor.SelectedObjects, sector.EntityRef) != -1

	if editor.BodiesVisible {
		for _, ibody := range sector.Bodies {
			mw.DrawBody(ibody)
		}
	}

	for _, segment := range sector.Segments {
		if segment.P == segment.Next.P {
			continue
		}

		segmentHovering := state.IndexOf(editor.HoveringObjects, segment) != -1
		segmentSelected := state.IndexOf(editor.SelectedObjects, segment) != -1

		if segment.AdjacentSector.Nil() {
			mw.Context.SetRGB(1, 1, 1)
		} else {
			mw.Context.SetRGB(1, 1, 0)
		}

		if sectorHovering || sectorSelected {
			if segment.AdjacentSector.Nil() {
				mw.Context.SetRGB(ColorSelectionPrimary[0], ColorSelectionPrimary[1], ColorSelectionPrimary[2])
			} else {
				mw.Context.SetRGB(ColorSelectionSecondary[0], ColorSelectionSecondary[1], ColorSelectionSecondary[2])
			}
		} else if segmentHovering {
			mw.Context.SetRGB(ColorSelectionSecondary[0], ColorSelectionSecondary[1], ColorSelectionSecondary[2])
		} else if segmentSelected {
			mw.Context.SetRGB(ColorSelectionPrimary[0], ColorSelectionPrimary[1], ColorSelectionPrimary[2])
		}

		// Highlight PVS sectors...
		for _, obj := range editor.SelectedObjects {
			switch selected := obj.(type) {
			case *concepts.EntityRef:
				if s2 := core.SectorFromDb(selected); s2 != nil && sector != s2 {
					if s2.PVSBody[sector.Entity] != nil {
						mw.Context.SetRGB(ColorPVS[0], ColorPVS[1], ColorPVS[2])
					}
				}

			}
		}

		// Draw segment
		mw.Context.SetLineWidth(1)
		mw.Context.NewSubPath()
		mw.Context.MoveTo(segment.P[0], segment.P[1])
		mw.Context.LineTo(segment.Next.P[0], segment.Next.P[1])
		mw.Context.ClosePath()
		mw.Context.Stroke()
		// Draw normal
		mw.Context.NewSubPath()
		ns := segment.Next.P.Add(&segment.P).Mul(0.5)
		ne := ns.Add(segment.Normal.Mul(4.0))
		mw.Context.MoveTo(ns[0], ns[1])
		mw.Context.LineTo(ne[0], ne[1])
		mw.Context.ClosePath()
		mw.Context.Stroke()

		if segmentSelected {
			mw.Context.SetRGB(ColorSelectionPrimary[0], ColorSelectionPrimary[1], ColorSelectionPrimary[2])
			mw.DrawHandle(&segment.P)
		} else if segmentHovering {
			mw.Context.SetRGB(ColorSelectionSecondary[0], ColorSelectionSecondary[1], ColorSelectionSecondary[2])
			mw.DrawHandle(&segment.P)
		} else {
			mw.Context.DrawRectangle(segment.P[0]-1, segment.P[1]-1, 2, 2)
			mw.Context.Stroke()
		}
	}

	if editor.SectorTypesVisible {
		text := reflect.TypeOf(sector).String()
		mw.Context.Push()
		mw.Context.SetRGB(0.3, 0.3, 0.3)
		mw.Context.DrawStringAnchored(text, sector.Center[0], sector.Center[1], 0.5, 0.5)
		mw.Context.Pop()
	}

	mw.Context.Pop()
}
