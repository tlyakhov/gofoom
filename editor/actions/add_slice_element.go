// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"reflect"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

// TODO: Expand this to also delete elements and move them
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
	a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
}
func (a *AddSliceElement) Redo() {
	a.State().Lock.Lock()
	defer a.State().Lock.Unlock()

	// SlicePtr is something like: *[]<some type>
	sliceElem := a.SlicePtr.Elem()
	s := reflect.Append(sliceElem, reflect.Zero(sliceElem.Type().Elem()))
	sliceElem.Set(s)
	newValue := sliceElem.Index(sliceElem.Len() - 1)
	if newValue.Kind() == reflect.Pointer {
		newValue.Set(reflect.New(newValue.Type().Elem()))
	}

	if serializable, ok := newValue.Interface().(ecs.Serializable); ok {
		serializable.OnAttach(a.State().ECS)
		serializable.Construct(nil)
	} else if subSerializable, ok := newValue.Interface().(ecs.SubSerializable); ok {
		subSerializable.Construct(a.State().ECS, nil)
	}
	a.State().ECS.ActAllControllers(ecs.ControllerRecalculate)
}
