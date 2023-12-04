package actions

import (
	"reflect"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"github.com/gotk3/gotk3/gdk"
)

type AddSliceElement struct {
	state.IEditor

	SlicePtr reflect.Value
}

func (a *AddSliceElement) Act() {
	a.Redo()
	a.ActionFinished(false)
}
func (a *AddSliceElement) Cancel()                             {}
func (a *AddSliceElement) Frame()                              {}
func (a *AddSliceElement) OnMouseDown(button *gdk.EventButton) {}
func (a *AddSliceElement) OnMouseMove()                        {}
func (a *AddSliceElement) OnMouseUp()                          {}

func (a *AddSliceElement) Undo() {
	if a.SlicePtr.Elem().Len() > 0 {
		a.SlicePtr.Elem().Slice(0, a.SlicePtr.Len()-1)
	}
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
}
func (a *AddSliceElement) Redo() {
	// SlicePtr is something like: *[]<some type>
	sliceElem := a.SlicePtr.Elem()
	newValue := reflect.Zero(sliceElem.Type().Elem())
	s := reflect.Append(sliceElem, newValue)
	sliceElem.Set(s)
	// Add more types?
	switch target := newValue.Interface().(type) {
	case core.Trigger:
		target.Construct(a.State().DB, nil)
	}
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
}
