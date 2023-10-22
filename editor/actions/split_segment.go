package actions

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"github.com/rs/xid"

	"tlyakhov/gofoom/core"

	"github.com/gotk3/gotk3/gdk"
)

type segmentSplitter struct {
	segments *[]*core.Segment
	index    int
	split    *core.Segment
	added    *core.Segment
}

type SplitSegment struct {
	state.IEditor

	NewSegments []*segmentSplitter
}

func (a *SplitSegment) OnMouseDown(button *gdk.EventButton) {}
func (a *SplitSegment) OnMouseMove()                        {}
func (a *SplitSegment) Frame()                              {}
func (a *SplitSegment) Act()                                {}

func (a *SplitSegment) Split(ss *segmentSplitter) bool {
	md := a.WorldGrid(&a.State().MouseDownWorld)
	m := a.WorldGrid(&a.State().MouseWorld)
	isect := new(concepts.Vector2)
	exists := ss.split.Intersect2D(md, m, isect)

	if !exists || *isect == ss.split.P || *isect == ss.split.Next.P {
		return false
	}

	copied := &core.Segment{}
	copied.SetParent(ss.split.Sector)
	copied.Construct(ss.split.Serialize())
	ss.added = copied
	ss.added.Name = xid.New().String()
	ss.added.P = *isect
	*ss.segments = append(*ss.segments, nil)
	copy((*ss.segments)[ss.index+1:], (*ss.segments)[ss.index:])
	(*ss.segments)[ss.index] = ss.added
	ss.added.Sector.Physical().Recalculate()
	a.NewSegments = append(a.NewSegments, ss)
	return true
}

func (a *SplitSegment) OnMouseUp() {
	a.NewSegments = []*segmentSplitter{}

	// Split only selected if any, otherwise all sectors/segments.
	all := a.State().SelectedObjects
	if len(all) == 0 || (len(all) == 1 && all[0] == a.State().World.Map) {
		all = make([]concepts.ISerializable, len(a.State().World.Sectors))
		i := 0
		for _, s := range a.State().World.Sectors {
			all[i] = s
			i++
		}
	}

	for _, obj := range all {
		if sector, ok := obj.(core.AbstractSector); ok {
			for j := 0; j < len(sector.Physical().Segments); j++ {
				if a.Split(&segmentSplitter{
					segments: &sector.Physical().Segments,
					split:    sector.Physical().Segments[j],
					index:    j + 1}) {
					j++ // Avoid infinite splitting.
				}
			}
		} else if _, ok := obj.(state.MapPoint); ok {
			// TODO...
		}
	}
	a.State().Modified = true
	a.ActionFinished(false)
}

func (a *SplitSegment) Cancel() {
	a.ActionFinished(true)
}

func (a *SplitSegment) Undo() {
	for _, ss := range a.NewSegments {
		reset := (*ss.segments)[:0]
		for _, seg := range *ss.segments {
			if seg.Name != ss.added.Name {
				reset = append(reset, seg)
			}
		}
		*ss.segments = reset
		ss.added.Sector.Physical().Recalculate()
	}
}
func (a *SplitSegment) Redo() {
	for _, ss := range a.NewSegments {
		*ss.segments = append(*ss.segments, nil)
		copy((*ss.segments)[ss.index+1:], (*ss.segments)[ss.index:])
		(*ss.segments)[ss.index] = ss.added
		ss.added.Sector.Physical().Recalculate()
	}
}
