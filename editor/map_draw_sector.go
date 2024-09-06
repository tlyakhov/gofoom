// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"

	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
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
	segmentHovering := editor.HoveringObjects.Contains(selection.SelectableFromInternalSegment(segment))
	segmentSelected := editor.SelectedObjects.Contains(selection.SelectableFromInternalSegment(segment))
	aHovering := segmentHovering || editor.HoveringObjects.Contains(selection.SelectableFromInternalSegmentA(segment))
	aSelected := segmentSelected || editor.SelectedObjects.Contains(selection.SelectableFromInternalSegmentA(segment))
	bHovering := segmentHovering || editor.HoveringObjects.Contains(selection.SelectableFromInternalSegmentB(segment))
	bSelected := segmentSelected || editor.SelectedObjects.Contains(selection.SelectableFromInternalSegmentB(segment))

	if segmentHovering {
		mw.Context.SetStrokeStyle(PatternSelectionSecondary)
	} else if segmentSelected {
		mw.Context.SetStrokeStyle(PatternSelectionPrimary)
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
		mw.Context.SetStrokeStyle(PatternSelectionPrimary)
		mw.DrawHandle(segment.A)
	} else if aSelected {
		mw.Context.SetStrokeStyle(PatternSelectionSecondary)
		mw.DrawHandle(segment.A)
	} else {
		mw.Context.SetRGB(1, 1, 1)
		mw.Context.DrawRectangle(segment.A[0]-1, segment.A[1]-1, 2, 2)
		mw.Context.Stroke()
	}

	if bHovering {
		mw.Context.SetStrokeStyle(PatternSelectionSecondary)
		mw.DrawHandle(segment.B)
	} else if bSelected {
		mw.Context.SetStrokeStyle(PatternSelectionPrimary)
		mw.DrawHandle(segment.B)
	} else {
		mw.Context.SetRGB(1, 1, 1)
		mw.Context.DrawRectangle(segment.B[0]-1, segment.B[1]-1, 2, 2)
		mw.Context.Stroke()
	}
	mw.Context.Pop()
}

func (mw *MapWidget) DrawSector(sector *core.Sector, isPartOfPVS bool) {
	if len(sector.Segments) == 0 {
		return
	}

	start := editor.ScreenToWorld(new(concepts.Vector2))
	end := editor.ScreenToWorld(&editor.Size)
	if !concepts.Vector2AABBIntersect(sector.Min.To2D(), sector.Max.To2D(), start, end, true) {
		return
	}

	mw.Context.Push()

	sectorHovering := editor.HoveringObjects.Contains(selection.SelectableFromSector(sector))
	sectorSelected := editor.SelectedObjects.Contains(selection.SelectableFromSector(sector))

	for _, segment := range sector.Segments {
		if segment.Next == nil || segment.P == segment.Next.P {
			continue
		}

		segmentHovering := editor.HoveringObjects.ContainsGrouped(selection.SelectableFromSegment(segment))
		segmentSelected := editor.SelectedObjects.ContainsGrouped(selection.SelectableFromSegment(segment))

		if segment.AdjacentSector == 0 {
			mw.Context.SetRGB(1, 1, 1)
		} else if segment.PortalTeleports {
			mw.Context.SetRGB(1, 0.5, 0)
		} else {
			mw.Context.SetRGB(1, 1, 0)
		}

		if sectorHovering || sectorSelected {
			if segment.AdjacentSector == 0 {
				mw.Context.SetStrokeStyle(PatternSelectionPrimary)
			} else {
				mw.Context.SetStrokeStyle(PatternSelectionSecondary)
			}
		} else if segmentHovering {
			mw.Context.SetStrokeStyle(PatternSelectionSecondary)
		} else if segmentSelected {
			mw.Context.SetStrokeStyle(PatternSelectionPrimary)
		}
		if isPartOfPVS && !segmentSelected && !sectorSelected {
			mw.Context.SetDash(4, 4)
		} else {
			mw.Context.SetDash()
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

			if segment.AdjacentSegment == nil || segment.PortalHasMaterial {
				img := editor.EntityImage(segment.Surface.Material, false)
				mw.Context.Push()
				ne = ns.Add(segment.Normal.Mul(-20.0))
				mw.Context.ScaleAbout(0.3, 0.3, ne[0], ne[1])
				mw.Context.DrawImageAnchored(img, (int)(ne[0]), (int)(ne[1]), 0.5, 0.5)
				mw.Context.Pop()
			}
		}

		if segmentSelected {
			mw.Context.SetStrokeStyle(PatternSelectionPrimary)
			mw.DrawHandle(&segment.P)
		} else if segmentHovering {
			mw.Context.SetStrokeStyle(PatternSelectionSecondary)
			mw.DrawHandle(&segment.P)
		} else {
			mw.Context.DrawRectangle(segment.P[0]-1, segment.P[1]-1, 2, 2)
			mw.Context.Stroke()
		}
	}

	if editor.SectorTypesVisible {
		text := sector.Entity.NameString(editor.ECS)
		mw.Context.Push()
		mw.Context.SetRGB(0.3, 0.3, 0.3)
		mw.Context.DrawStringAnchored(text, sector.Center[0], sector.Center[1], 0.5, 0.5)
		mw.Context.Pop()
	}

	mw.Context.Pop()
}

func (mw *MapWidget) DrawPath(waypoint *behaviors.ActionWaypoint) {
	/*start := editor.ScreenToWorld(new(concepts.Vector2))
	end := editor.ScreenToWorld(&editor.Size)
	if !concepts.Vector2AABBIntersect(path.Min.To2D(), path.Max.To2D(), start, end, true) {
		return
	}*/

	mw.Context.Push()

	waypointHovering := editor.HoveringObjects.ContainsGrouped(selection.SelectableFromActionWaypoint(waypoint))
	waypointSelected := editor.SelectedObjects.ContainsGrouped(selection.SelectableFromActionWaypoint(waypoint))

	if waypointSelected {
		mw.Context.SetStrokeStyle(PatternSelectionPrimary)
		mw.DrawHandle(waypoint.P.To2D())
	} else if waypointHovering {
		mw.Context.SetStrokeStyle(PatternSelectionSecondary)
		mw.DrawHandle(waypoint.P.To2D())
	} else {
		mw.Context.DrawRectangle(waypoint.P[0]-1, waypoint.P[1]-1, 2, 2)
		mw.Context.Stroke()
	}

	transition := behaviors.GetActionTransition(waypoint.ECS, waypoint.Entity)
	if transition == nil || transition.Next == 0 {
		return
	}

	next := behaviors.GetActionWaypoint(waypoint.ECS, transition.Next)

	if next == nil {
		return
	}

	mw.Context.SetRGB(0.4, 1, 0.6)

	if waypointHovering || waypointSelected {
		mw.Context.SetStrokeStyle(PatternSelectionPrimary)
	} else if waypointHovering {
		mw.Context.SetStrokeStyle(PatternSelectionSecondary)
	} else if waypointSelected {
		mw.Context.SetStrokeStyle(PatternSelectionPrimary)
	}

	// Draw segment
	mw.Context.SetDash(3, 8)
	mw.Context.SetLineWidth(1)
	mw.Context.NewSubPath()
	mw.Context.MoveTo(waypoint.P[0], waypoint.P[1])
	mw.Context.LineTo(next.P[0], next.P[1])
	mw.Context.ClosePath()
	mw.Context.Stroke()
	mw.Context.SetDash()

	if waypointHovering || waypointSelected {
		normal := &concepts.Vector3{next.P[1] - waypoint.P[1], waypoint.P[0] - next.P[0], 0}
		normal.NormSelf()
		ns := next.P.Add(&waypoint.P).Mul(0.5)
		ne := ns.Add(normal.Mul(20.0))
		txtLength := fmt.Sprintf("%.0f", next.P.Sub(&waypoint.P).Length())
		mw.Context.DrawStringAnchored(txtLength, ne[0], ne[1], 0.5, 0.5)
	}

	/*if editor.SectorTypesVisible {
		text := path.Entity.NameString(editor.ECS)
		mw.Context.Push()
		mw.Context.SetRGB(0.3, 0.3, 0.3)
		mw.Context.DrawStringAnchored(text, path.Center[0], path.Center[1], 0.5, 0.5)
		mw.Context.Pop()
	}*/

	mw.Context.Pop()
}
