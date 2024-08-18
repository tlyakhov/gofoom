// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"reflect"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

type AddSliceElement struct {
	state.IEditor
	Parent   any
	SlicePtr reflect.Value
	Concrete reflect.Type
}

func (a *AddSliceElement) Act() {
	a.Redo()
	a.ActionFinished(false, true, false)
}

func (a *AddSliceElement) Undo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()
	if a.SlicePtr.Elem().Len() > 0 {
		a.SlicePtr.Elem().Slice(0, a.SlicePtr.Len()-1)
	}
	a.State().DB.ActAllControllers(ecs.ControllerRecalculate)
}
func (a *AddSliceElement) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	// SlicePtr is something like: *[]<some type>
	sliceElem := a.SlicePtr.Elem()
	s := reflect.Append(sliceElem, reflect.Zero(sliceElem.Type().Elem()))
	sliceElem.Set(s)
	newValue := sliceElem.Index(sliceElem.Len() - 1).Addr()
	// Add more types?
	switch target := newValue.Interface().(type) {
	case **core.Script:
		*target = new(core.Script)
	case **materials.ShaderStage:
		*target = new(materials.ShaderStage)
	case **materials.Sprite:
		*target = new(materials.Sprite)
	case **behaviors.InventorySlot:
		*target = new(behaviors.InventorySlot)
	}
	if serializable, ok := newValue.Elem().Interface().(ecs.Serializable); ok {
		serializable.SetECS(a.State().DB)
		serializable.Construct(nil)
	}
	a.State().DB.ActAllControllers(ecs.ControllerRecalculate)
}
