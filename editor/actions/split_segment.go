// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
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

func (a *SplitSegment) Activate() {}

func (a *SplitSegment) split(ss *segmentSplitter) bool {
	md := a.WorldGrid(&a.State().MouseDownWorld)
	m := a.WorldGrid(&a.State().MouseWorld)
	isect := new(concepts.Vector2)
	u := ss.original.Intersect2D(md, m, isect)

	if u < 0 || *isect == ss.original.P.Render || *isect == ss.original.Next.P.Render {
		return false
	}

	ss.added = ss.original.Split(*isect)
	ss.added.P.SetAll(*isect)
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
	if a.State().Selection.Empty() {
		arena := ecs.ArenaFor[core.Sector](core.SectorCID)
		segments = make(containers.Set[*core.SectorSegment])
		for i := range arena.Cap() {
			if sector := arena.Value(i); sector != nil {
				segments.AddAll(sector.Segments...)
			}
		}
	} else {
		segments = make(containers.Set[*core.SectorSegment])
		for _, s := range a.State().Selection.Exact {
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
		a.split(&segmentSplitter{original: seg})
	}
	a.State().Modified = true
	a.ActionFinished(false, true, true)
	return true
}

func (a *SplitSegment) Cancel() {
	a.ActionFinished(true, true, true)
}

func (a *SplitSegment) Status() string {
	return ""
}
