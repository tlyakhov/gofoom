package actions

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/components/core"

	"fyne.io/fyne/v2/driver/desktop"
)

type segmentSplitter struct {
	original *core.Segment
	added    *core.Segment
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
	all := a.State().SelectedObjects
	if len(all) == 0 || (len(all) == 1 && all[0] == a.State().DB) {
		allSectors := a.State().DB.Components[core.SectorComponentIndex]
		all = make([]any, len(allSectors))
		i := 0
		for _, s := range allSectors {
			all[i] = s.Ref()
			i++
		}
	}

	for _, selected := range all {
		switch target := selected.(type) {
		case *concepts.EntityRef:
			if sector := core.SectorFromDb(target); sector != nil {
				for j := 0; j < len(sector.Segments); j++ {
					if a.Split(&segmentSplitter{
						original: sector.Segments[j]}) {
						j++ // Avoid infinite splitting.
					}
				}
			}
		case *core.Segment:
			a.Split(&segmentSplitter{original: target})
		}

	}
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}

func (a *SplitSegment) Cancel() {
	a.ActionFinished(true, true, true)
}

func (a *SplitSegment) Undo() {
	for _, ss := range a.NewSegments {
		reset := make([]*core.Segment, 0)
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
