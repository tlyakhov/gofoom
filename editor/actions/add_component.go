package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2/driver/desktop"
)

type AddComponent struct {
	state.IEditor

	Entities []uint64
	Index    int
}

func (a *AddComponent) Act() {
	a.Redo()
	a.ActionFinished(false, true, a.Index == core.SectorComponentIndex)
}
func (a *AddComponent) Cancel()                             {}
func (a *AddComponent) Frame()                              {}
func (a *AddComponent) OnMouseDown(evt *desktop.MouseEvent) {}
func (a *AddComponent) OnMouseMove()                        {}
func (a *AddComponent) OnMouseUp()                          {}

func (a *AddComponent) Undo() {
	for _, entity := range a.Entities {
		a.State().DB.Detach(a.Index, entity)
	}
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
}
func (a *AddComponent) Redo() {
	for _, entity := range a.Entities {
		a.State().DB.NewComponent(entity, a.Index)
	}
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
}

func (a *AddComponent) RequiresLock() bool { return true }
