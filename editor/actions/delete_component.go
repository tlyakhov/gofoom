package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2/driver/desktop"
)

type DeleteComponent struct {
	state.IEditor

	Components      map[uint64]concepts.Attachable
	SectorForEntity map[uint64]*core.Sector
}

func (a *DeleteComponent) Act() {
	a.SectorForEntity = make(map[uint64]*core.Sector)
	a.Redo()
	a.ActionFinished(false, true, false)
}
func (a *DeleteComponent) Cancel()                             {}
func (a *DeleteComponent) Frame()                              {}
func (a *DeleteComponent) OnMouseDown(evt *desktop.MouseEvent) {}
func (a *DeleteComponent) OnMouseMove()                        {}
func (a *DeleteComponent) OnMouseUp()                          {}

func (a *DeleteComponent) Undo() {
	for entity, component := range a.Components {
		a.State().DB.AttachTyped(entity, component)
		switch target := component.(type) {
		case *core.Body:
			entity := component.Ref().Entity
			if a.SectorForEntity[entity] != nil {
				a.SectorForEntity[entity].Bodies[entity] = target.EntityRef
			}
		}
	}
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
}
func (a *DeleteComponent) Redo() {
	for _, component := range a.Components {
		// TODO: Save material references
		switch target := component.(type) {
		case *core.Body:
			entity := component.Ref().Entity
			if target.SectorEntityRef != nil {
				a.SectorForEntity[entity] = target.Sector()
				delete(a.SectorForEntity[entity].Bodies, entity)
			}
		}
		a.State().DB.DetachByType(component)
	}
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
}
