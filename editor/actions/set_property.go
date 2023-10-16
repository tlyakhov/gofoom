package actions

import (
	"fmt"
	"reflect"

	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gdk"
)

type SetProperty struct {
	state.IEditor
	*state.PropertyGridField
	Original []reflect.Value
	ToSet    reflect.Value
}

func (a *SetProperty) OnMouseDown(button *gdk.EventButton) {}
func (a *SetProperty) OnMouseMove()                        {}
func (a *SetProperty) OnMouseUp()                          {}
func (a *SetProperty) Cancel()                             {}
func (a *SetProperty) Frame()                              {}

func (a *SetProperty) Act() {
	for _, v := range a.Values {
		origValue := reflect.ValueOf(v.Elem().Interface())
		a.Original = append(a.Original, origValue)
		if a.Source.Name == "ID" {
			// IDs are special, because we have to also update the containing map key.
			a.ParentCollection.SetMapIndex(origValue, reflect.Value{})
			a.ParentCollection.SetMapIndex(a.ToSet, reflect.ValueOf(a.Parent))
		}
		v.Elem().Set(a.ToSet)
	}
	a.State().Modified = true
	a.ActionFinished(false)
}

func (a *SetProperty) Undo() {
	for i, v := range a.Values {
		fmt.Printf("Undo: %v\n", a.Original[i].String())
		v.Elem().Set(a.Original[i])
	}
}
func (a *SetProperty) Redo() {
	for _, v := range a.Values {
		fmt.Printf("Redo: %v\n", a.ToSet.String())
		v.Elem().Set(a.ToSet)
	}
}
