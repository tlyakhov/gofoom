// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"slices"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/ecs"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"
)

type segmentSplitter struct {
	original *core.SectorSegment
	added    *core.SectorSegment
}

type SplitSegment struct {
	Place

	NewSegments []*segmentSplitter
}

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

func (a *SplitSegment) EndPoint() bool {
	if !a.Place.EndPoint() {
		return false
	}
	a.NewSegments = []*segmentSplitter{}

	// Split only selected if any, otherwise all sectors/segments.
	// TODO: also split internal segments
	var segments containers.Set[*core.SectorSegment]
	if a.State().SelectedObjects.Empty() {
		col := ecs.ColumnFor[core.Sector](a.State().ECS, core.SectorCID)
		segments = make(containers.Set[*core.SectorSegment])
		for i := range col.Cap() {
			if sector := col.Value(i); sector != nil {
				segments.AddAll(sector.Segments...)
			}
		}
	} else {
		segments = make(containers.Set[*core.SectorSegment])
		for _, s := range a.State().SelectedObjects.Exact {
			switch s.Type {
			// Segments:
			case selection.SelectableLow, selection.SelectableMid,
				selection.SelectableHi, selection.SelectableSectorSegment:
				segments.Add(s.SectorSegment)
			// Sectors:
			case selection.SelectableCeiling, selection.SelectableFloor,
				selection.SelectableSector:
				for _, seg := range s.Sector.Segments {
					segments.Add(seg)
				}
			}
		}
	}

	for seg := range segments {
		a.Split(&segmentSplitter{original: seg})
	}
	a.State().Modified = true
	a.ActionFinished(false, true, true)
	return true
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
