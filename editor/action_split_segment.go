package main

import (
	"github.com/rs/xid"
	"github.com/tlyakhov/gofoom/concepts"

	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/core"
)

type SplitSegment struct {
	segments *[]*core.Segment
	index    int
	split    *core.Segment
	added    *core.Segment
}

type SplitSegmentAction struct {
	*Editor
	NewSegments []*SplitSegment
}

func (a *SplitSegmentAction) OnMouseDown(button *gdk.EventButton) {}
func (a *SplitSegmentAction) OnMouseMove()                        {}
func (a *SplitSegmentAction) Frame()                              {}
func (a *SplitSegmentAction) Act()                                {}

func (a *SplitSegmentAction) Split(ss *SplitSegment) bool {
	md := a.WorldGrid(a.MouseDownWorld)
	m := a.WorldGrid(a.MouseWorld)
	isect, exists := ss.split.Intersect2D(md, m)

	if !exists || isect == ss.split.A || isect == ss.split.B {
		return false
	}

	// TODO: in the future this should be serialize->deserialize for a good clone.
	copied := *ss.split
	ss.added = &copied
	ss.added.ID = xid.New().String()
	ss.added.A = isect
	*ss.segments = append(*ss.segments, nil)
	copy((*ss.segments)[ss.index+1:], (*ss.segments)[ss.index:])
	(*ss.segments)[ss.index] = ss.added
	ss.added.Sector.Physical().Recalculate()
	a.NewSegments = append(a.NewSegments, ss)
	return true
}

func (a *SplitSegmentAction) OnMouseUp() {
	a.NewSegments = []*SplitSegment{}

	// Split only selected if any, otherwise all sectors/segments.
	all := a.SelectedObjects
	if all == nil || len(all) == 0 || (len(all) == 1 && all[0] == a.GameMap.Map) {
		all = make([]concepts.ISerializable, len(a.GameMap.Sectors))
		i := 0
		for _, s := range a.GameMap.Sectors {
			all[i] = s
			i++
		}
	}

	for _, obj := range all {
		if sector, ok := obj.(core.AbstractSector); ok {
			for j := 0; j < len(sector.Physical().Segments); j++ {
				if a.Split(&SplitSegment{
					segments: &sector.Physical().Segments,
					split:    sector.Physical().Segments[j],
					index:    j + 1}) {
					j++ // Avoid infinite splitting.
				}
			}
		} else if _, ok := obj.(MapPoint); ok {
			// TODO...
		}
	}
	a.ActionFinished(false)
}

func (a *SplitSegmentAction) Cancel() {
	a.ActionFinished(true)
}

func (a *SplitSegmentAction) Undo() {
	for _, ss := range a.NewSegments {
		reset := (*ss.segments)[:0]
		for _, seg := range *ss.segments {
			if seg.ID != ss.added.ID {
				reset = append(reset, seg)
			}
		}
		*ss.segments = reset
		ss.added.Sector.Physical().Recalculate()
	}
}
func (a *SplitSegmentAction) Redo() {
	for _, ss := range a.NewSegments {
		*ss.segments = append(*ss.segments, nil)
		copy((*ss.segments)[ss.index+1:], (*ss.segments)[ss.index:])
		(*ss.segments)[ss.index] = ss.added
		ss.added.Sector.Physical().Recalculate()
	}
}
