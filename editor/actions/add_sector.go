package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gdk"
)

type AddSector struct {
	state.IEditor

	Mode   string
	Sector *core.Sector
}

func (a *AddSector) Act() {
	a.SetMapCursor("crosshair")
	a.Mode = "AddSector"
	a.SelectObjects([]any{a.Sector})
	//set cursor
}

func (a *AddSector) Cancel() {
	a.RemoveFromMap()
	a.Sector.Segments = []*core.Segment{}
	a.SelectObjects([]any{})
	a.ActionFinished(true)
}

func (a *AddSector) RemoveFromMap() {
	// remove sector
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
}

func (a *AddSector) AddToMap() {
	// Add sector
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
}

func (a *AddSector) OnMouseDown(button *gdk.EventButton) {
	a.Mode = "AddSectorSegment"

	seg := core.Segment{}
	seg.Construct(nil)
	seg.Sector = a.Sector
	seg.HiMaterial = controllers.DefaultMaterial(a.State().DB)
	seg.LoMaterial = controllers.DefaultMaterial(a.State().DB)
	seg.MidMaterial = controllers.DefaultMaterial(a.State().DB)
	seg.P = *a.WorldGrid(&a.State().MouseDownWorld)

	segs := a.Sector.Segments
	if len(segs) > 0 {
		seg.Prev = segs[len(segs)-1]
		seg.Next = segs[0]
		seg.Next.Prev = &seg
		seg.Prev.Next = &seg
	}

	a.Sector.Segments = append(segs, &seg)
	a.AddToMap()
}
func (a *AddSector) OnMouseMove() {
	if a.Mode != "AddSectorSegment" {
		return
	}

	segs := a.Sector.Segments
	seg := segs[len(segs)-1]
	seg.P = *a.WorldGrid(&a.State().MouseWorld)
}

func (a *AddSector) OnMouseUp() {
	a.Mode = "AddSector"

	segs := a.Sector.Segments
	if len(segs) > 1 {
		first := segs[0]
		last := segs[len(segs)-1]
		if last.P.Sub(&first.P).Length() < state.SegmentSelectionEpsilon {
			a.Sector.Segments = segs[:(len(segs) - 1)]
			a.State().Modified = true
			a.ActionFinished(false)
			// Recalculate
		}
	}
	// TODO: right-mouse button end
}

func (a *AddSector) Frame() {}

func (a *AddSector) Undo() {
	a.RemoveFromMap()
}
func (a *AddSector) Redo() {
	a.AddToMap()
}
