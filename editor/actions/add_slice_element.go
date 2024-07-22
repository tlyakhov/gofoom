// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"reflect"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"
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
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
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
		script := *target
		script.SetDB(a.State().DB)
		script.Construct(nil)
	case **materials.ShaderStage:
		*target = new(materials.ShaderStage)
		stage := *target
		stage.SetDB(a.State().DB)
		stage.Construct(nil)
	case **materials.Sprite:
		*target = new(materials.Sprite)
		sprite := *target
		sprite.SetDB(a.State().DB)
		sprite.Construct(nil)
		/*case *concepts.IAnimation:
		reflect.ValueOf(target).Elem().Set(reflect.New(a.Concrete))
		anim := *target
		anim.SetDB(a.State().DB)
		anim.Construct(nil)
		a.State().DB.Simulation.AttachAnimation(anim.GetName(), anim)*/
	}
	a.State().DB.ActAllControllers(concepts.ControllerRecalculate)
}
