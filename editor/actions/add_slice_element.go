package actions

import (
	"reflect"
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
	s := reflect.Append(a.SlicePtr.Elem(), reflect.Zero(a.SlicePtr.Elem().Type().Elem()))
	a.SlicePtr.Elem().Set(s)
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
}
