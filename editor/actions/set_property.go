package actions

import (
	"fmt"
	"reflect"

	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gdk"
)

type SetProperty struct {
	state.IEditor

	Fields   []reflect.Value
	Original []reflect.Value
	ToSet    reflect.Value
}

func (a *SetProperty) OnMouseDown(button *gdk.EventButton) {}
func (a *SetProperty) OnMouseMove()                        {}
func (a *SetProperty) OnMouseUp()                          {}
func (a *SetProperty) Cancel()                             {}
func (a *SetProperty) Frame()                              {}

func (a *SetProperty) Act() {
	for _, field := range a.Fields {
		a.Original = append(a.Original, reflect.ValueOf(field.Elem().Interface()))
		field.Elem().Set(a.ToSet)
	}
	a.State().Modified = true
	a.ActionFinished(false)
}

func (a *SetProperty) Undo() {
	for i, field := range a.Fields {
		fmt.Println(a.Original[i].String())
		field.Elem().Set(a.Original[i])
	}
}
func (a *SetProperty) Redo() {
	for _, field := range a.Fields {
		fmt.Println(a.ToSet.String())
		field.Elem().Set(a.ToSet)
	}
}
