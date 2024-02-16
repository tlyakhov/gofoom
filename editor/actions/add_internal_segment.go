package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/controllers"

	"fyne.io/fyne/v2/driver/desktop"
)

type AddInternalSegment struct {
	AddBody
	*core.InternalSegment
}

func (a *AddInternalSegment) Act() {
	a.SetMapCursor(desktop.CrosshairCursor)
	a.Mode = "AddInternalSegment"
	a.SelectObjects([]any{a.EntityRef}, true)
	//set cursor
}

func (a *AddInternalSegment) OnMouseDown(evt *desktop.MouseEvent) {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	switch a.Mode {
	case "AddInternalSegment":
		a.Mode = "AddInternalSegmentA"
		a.Surface.Material = controllers.DefaultMaterial(a.State().DB)
		a.AttachToSector()
	case "AddInternalSegmentA":
	}
}
func (a *AddInternalSegment) OnMouseMove() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	worldGrid := a.WorldGrid(&a.State().MouseWorld)

	for _, ref := range a.State().DB.All(core.SectorComponentIndex) {
		sector := ref.(*core.Sector)
		if sector.IsPointInside2D(worldGrid) {
			a.ContainingSector = sector
			break
		}
	}

	a.DetachFromSector()
	a.AttachToSector()
	switch a.Mode {
	case "AddInternalSegment":
		fallthrough
	case "AddInternalSegmentA":
		*a.A = *worldGrid
		fallthrough
	case "AddInternalSegmentB":
		*a.B = *worldGrid
	}
	if a.ContainingSector != nil {
		a.Bottom, a.Top = a.ContainingSector.SlopedZOriginal(worldGrid)
	}
	a.Recalculate()
}

func (a *AddInternalSegment) OnMouseUp() {
	switch a.Mode {
	case "AddInternalSegmentA":
		*a.A = *a.WorldGrid(&a.State().MouseWorld)
		a.Mode = "AddInternalSegmentB"
	case "AddInternalSegmentB":
		*a.B = *a.WorldGrid(&a.State().MouseWorld)
		a.Mode = "AddInternalSegment"
		a.State().Lock.Lock()
		a.State().Modified = true
		a.Recalculate()
		a.State().Lock.Unlock()
		a.ActionFinished(false, true, true)
	}
}
