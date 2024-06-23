// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
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

func (mw *MapWidget) DrawInternalSegment(segment *core.InternalSegment) {
	mw.Context.Push()
	segmentHovering := core.SelectableFromInternalSegment(segment).ExactIndexIn(editor.HoveringObjects) != -1
	segmentSelected := core.SelectableFromInternalSegment(segment).ExactIndexIn(editor.SelectedObjects) != -1
	aHovering := segmentHovering || core.SelectableFromInternalSegmentA(segment).ExactIndexIn(editor.HoveringObjects) != -1
	aSelected := segmentSelected || core.SelectableFromInternalSegmentA(segment).ExactIndexIn(editor.SelectedObjects) != -1
	bHovering := segmentHovering || core.SelectableFromInternalSegmentB(segment).ExactIndexIn(editor.HoveringObjects) != -1
	bSelected := segmentSelected || core.SelectableFromInternalSegmentB(segment).ExactIndexIn(editor.SelectedObjects) != -1

	if segmentHovering {
		mw.Context.SetRGB(ColorSelectionSecondary[0], ColorSelectionSecondary[1], ColorSelectionSecondary[2])
	} else if segmentSelected {
		mw.Context.SetRGB(ColorSelectionPrimary[0], ColorSelectionPrimary[1], ColorSelectionPrimary[2])
	} else {
		mw.Context.SetRGB(1, 1, 1)
	}

	// Draw segment
	if segment.A.Dist2(segment.B) > state.SegmentSelectionEpsilon {
		mw.Context.SetLineWidth(1)
		mw.Context.NewSubPath()
		mw.Context.MoveTo(segment.A[0], segment.A[1])
		mw.Context.LineTo(segment.B[0], segment.B[1])
		mw.Context.ClosePath()
		mw.Context.Stroke()
		// Draw normal
		mw.Context.NewSubPath()
		ns := segment.B.Add(segment.A).Mul(0.5)
		ne := ns.Add(segment.Normal.Mul(4.0))
		mw.Context.MoveTo(ns[0], ns[1])
		mw.Context.LineTo(ne[0], ne[1])
		mw.Context.ClosePath()
		mw.Context.Stroke()
		if segmentHovering || segmentSelected {
			ne = ns.Add(segment.Normal.Mul(20.0))
			txtLength := fmt.Sprintf("%.0f", segment.Length)
			mw.Context.DrawStringAnchored(txtLength, ne[0], ne[1], 0.5, 0.5)
		}
	}

	if aHovering {
		mw.Context.SetRGB(ColorSelectionSecondary[0], ColorSelectionSecondary[1], ColorSelectionSecondary[2])
		mw.DrawHandle(segment.A)
	} else if aSelected {
		mw.Context.SetRGB(ColorSelectionPrimary[0], ColorSelectionPrimary[1], ColorSelectionPrimary[2])
		mw.DrawHandle(segment.A)
	} else {
		mw.Context.SetRGB(1, 1, 1)
		mw.Context.DrawRectangle(segment.A[0]-1, segment.A[1]-1, 2, 2)
		mw.Context.Stroke()
	}

	if bHovering {
		mw.Context.SetRGB(ColorSelectionSecondary[0], ColorSelectionSecondary[1], ColorSelectionSecondary[2])
		mw.DrawHandle(segment.B)
	} else if bSelected {
		mw.Context.SetRGB(ColorSelectionPrimary[0], ColorSelectionPrimary[1], ColorSelectionPrimary[2])
		mw.DrawHandle(segment.B)
	} else {
		mw.Context.SetRGB(1, 1, 1)
		mw.Context.DrawRectangle(segment.B[0]-1, segment.B[1]-1, 2, 2)
		mw.Context.Stroke()
	}
	mw.Context.Pop()
}

func (mw *MapWidget) DrawSector(sector *core.Sector) {
	if len(sector.Segments) == 0 {
		return
	}

	start := editor.ScreenToWorld(new(concepts.Vector2))
	end := editor.ScreenToWorld(&editor.Size)
	if !concepts.Vector2AABBIntersect(sector.Min.To2D(), sector.Max.To2D(), start, end, true) {
		return
	}

	mw.Context.Push()

	sectorHovering := core.SelectableFromSector(sector).IndexIn(editor.HoveringObjects) != -1
	sectorSelected := core.SelectableFromSector(sector).IndexIn(editor.SelectedObjects) != -1

	for _, segment := range sector.Segments {
		if segment.Next == nil || segment.P == segment.Next.P {
			continue
		}

		segmentHovering := core.SelectableFromSegment(segment).IndexIn(editor.HoveringObjects) != -1
		segmentSelected := core.SelectableFromSegment(segment).IndexIn(editor.SelectedObjects) != -1

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
			if obj.Sector == nil || sector == obj.Sector {
				continue
			}
			if obj.Sector.PVS[sector.Entity] != nil {
				mw.Context.SetRGB(ColorPVS[0], ColorPVS[1], ColorPVS[2])
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

		if sectorHovering || sectorSelected || segmentHovering || segmentSelected {
			ne = ns.Add(segment.Normal.Mul(20.0))
			txtLength := fmt.Sprintf("%.0f", segment.Length)
			mw.Context.DrawStringAnchored(txtLength, ne[0], ne[1], 0.5, 0.5)
		}

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

	if editor.BodiesVisible {
		for _, ref := range sector.Bodies {
			mw.DrawBody(ref)
		}
	}

	for _, ref := range sector.InternalSegments {
		mw.DrawInternalSegment(core.InternalSegmentFromDb(ref))
	}

	mw.Context.Pop()
}
