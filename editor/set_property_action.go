package main

import (
	"fmt"
	"reflect"

	"github.com/gotk3/gotk3/gdk"
)

type SetPropertyAction struct {
	*Editor
	Fields   []reflect.Value
	Original []reflect.Value
	ToSet    reflect.Value
}

func (a *SetPropertyAction) OnMouseDown(button *gdk.EventButton) {}
func (a *SetPropertyAction) OnMouseMove()                        {}
func (a *SetPropertyAction) OnMouseUp()                          {}
func (a *SetPropertyAction) Cancel()                             {}
func (a *SetPropertyAction) Frame()                              {}

func (a *SetPropertyAction) Act() {
	for _, field := range a.Fields {
		a.Original = append(a.Original, reflect.ValueOf(field.Elem().Interface()))
		field.Elem().Set(a.ToSet)
	}
	a.RefreshPropertyGrid()
}

func (a *SetPropertyAction) Undo() {
	for i, field := range a.Fields {
		fmt.Println(a.Original[i].String())
		field.Elem().Set(a.Original[i])
	}
	a.RefreshPropertyGrid()
}
func (a *SetPropertyAction) Redo() {
	for _, field := range a.Fields {
		fmt.Println(a.ToSet.String())
		field.Elem().Set(a.ToSet)
	}
	a.RefreshPropertyGrid()
}
