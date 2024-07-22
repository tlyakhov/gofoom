// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"slices"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/components/core"

	"fyne.io/fyne/v2/driver/desktop"
)

type segmentSplitter struct {
	original *core.SectorSegment
	added    *core.SectorSegment
}

type SplitSegment struct {
	state.IEditor

	NewSegments []*segmentSplitter
}

func (a *SplitSegment) OnMouseDown(evt *desktop.MouseEvent) {}
func (a *SplitSegment) OnMouseMove()                        {}

func (a *SplitSegment) Act() {}

func (a *SplitSegment) Split(ss *segmentSplitter) bool {
	md := a.WorldGrid(&a.State().MouseDownWorld)
	m := a.WorldGrid(&a.State().MouseWorld)
	isect := new(concepts.Vector2)
	exists := ss.original.Intersect2D(md, m, isect)

	if !exists || *isect == ss.original.P || *isect == ss.original.Next.P {
		return false
	}

	ss.added = ss.original.Split(*isect)
	ss.added.P = *isect
	ss.added.Sector.Recalculate()
	a.NewSegments = append(a.NewSegments, ss)
	return true
}

func (a *SplitSegment) OnMouseUp() {
	a.NewSegments = []*segmentSplitter{}

	// Split only selected if any, otherwise all sectors/segments.
	// TODO: also split internal segments
	var segments []*core.SectorSegment
	if a.State().SelectedObjects.Empty() {
		allSectors := a.State().DB.AllOfType(core.SectorComponentIndex)
		segments = make([]*core.SectorSegment, 0)
		for _, attachable := range allSectors {
			sector := attachable.(*core.Sector)
			segments = append(segments, sector.Segments...)
		}
	} else {
		segments = make([]*core.SectorSegment, 0)
		visited := make(map[*core.SectorSegment]bool)
		for _, s := range a.State().SelectedObjects.Exact {
			switch s.Type {
			// Segments:
			case core.SelectableLow:
				fallthrough
			case core.SelectableMid:
				fallthrough
			case core.SelectableHi:
				fallthrough
			case core.SelectableSectorSegment:
				segments = append(segments, s.SectorSegment)
				visited[s.SectorSegment] = true
			// Sectors:
			case core.SelectableCeiling:
				fallthrough
			case core.SelectableFloor:
				fallthrough
			case core.SelectableSector:
				for _, seg := range s.Sector.Segments {
					segments = append(segments, seg)
					visited[seg] = true
				}
			}
		}
	}

	for _, seg := range segments {
		a.Split(&segmentSplitter{original: seg})
	}
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}

func (a *SplitSegment) Cancel() {
	a.ActionFinished(true, true, true)
}

func (a *SplitSegment) Undo() {
	for _, ss := range a.NewSegments {
		reset := make([]*core.SectorSegment, 0)
		segments := ss.original.Sector.Segments
		for _, seg := range segments {
			if seg != ss.added {
				reset = append(reset, seg)
			}
		}
		ss.original.Sector.Segments = reset
		ss.added.Sector.Recalculate()
	}
}
func (a *SplitSegment) Redo() {
	for _, ss := range a.NewSegments {
		index := 0
		for i, seg := range ss.original.Sector.Segments {
			if seg == ss.original {
				index = i
				break
			}
		}
		ss.original.Sector.Segments = slices.Insert(ss.original.Sector.Segments, index+1, ss.added)
		ss.added.Sector.Recalculate()
	}
}

func (a *SplitSegment) RequiresLock() bool { return true }

func (a *SplitSegment) Status() string {
	return ""
}
