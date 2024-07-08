// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2/driver/desktop"
)

type DeleteComponent struct {
	state.IEditor

	Components      map[concepts.Entity]concepts.Attachable
	SectorForEntity map[concepts.Entity]*core.Sector
}

func (a *DeleteComponent) Act() {
	a.SectorForEntity = make(map[concepts.Entity]*core.Sector)
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
			entity := component.GetEntity()
			if a.SectorForEntity[entity] != nil {
				a.SectorForEntity[entity].Bodies[entity] = target
			}
		}
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
func (a *DeleteComponent) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	for _, component := range a.Components {
		// TODO: Save material references
		switch target := component.(type) {
		case *core.Body:
			entity := component.GetEntity()
			if target.SectorEntity != 0 {
				a.SectorForEntity[entity] = target.Sector()
				delete(a.SectorForEntity[entity].Bodies, entity)
			}
		}
		a.State().DB.DetachByType(component)
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}

func (a *DeleteComponent) Status() string {
	return ""
}
