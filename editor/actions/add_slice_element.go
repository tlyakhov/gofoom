package actions

import (
	"reflect"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"

	"github.com/gotk3/gotk3/gdk"
)

type AddSliceElement struct {
	state.IEditor
	Parent   any
	SlicePtr reflect.Value
}

func (a *AddSliceElement) Act() {
	a.Redo()
	a.ActionFinished(false, true, false)
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
	s := reflect.Append(sliceElem, reflect.Zero(sliceElem.Type().Elem()))
	sliceElem.Set(s)
	newValue := sliceElem.Index(sliceElem.Len() - 1).Addr()
	// Add more types?
	switch target := newValue.Interface().(type) {
	case *core.Script:
		target.Construct(a.State().DB, nil)
	case *materials.ShaderStage:
		target.Construct(a.Parent.(*materials.Shader), nil)
	}
	a.State().DB.NewControllerSet().ActGlobal(concepts.ControllerRecalculate)
}
