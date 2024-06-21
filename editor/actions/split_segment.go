// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
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
func (a *SplitSegment) Frame()                              {}
func (a *SplitSegment) Act()                                {}

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
	if len(a.State().SelectedObjects) == 0 {
		allSectors := a.State().DB.Components[core.SectorComponentIndex]
		segments = make([]*core.SectorSegment, 0)
		for _, attachable := range allSectors {
			sector := attachable.(*core.Sector)
			segments = append(segments, sector.Segments...)
		}
	} else {
		segments = make([]*core.SectorSegment, 0)
		visited := make(map[*core.SectorSegment]bool)
		for _, s := range a.State().SelectedObjects {
			switch s.Type {
			case state.SelectableSectorSegment:
				segments = append(segments, s.SectorSegment)
				visited[s.SectorSegment] = true
			case state.SelectableSector:
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
	panic("unimplemented redo:splitsegment")
	for _, ss := range a.NewSegments {
		// TODO
		ss.added.Sector.Recalculate()
	}
}

func (a *SplitSegment) RequiresLock() bool { return true }
