package actions

import (
	"fmt"
	"reflect"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
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

func (a *SetProperty) FireHooks() {
	// TODO: this has a bug - all these fire for the same parent. We need to
	// be able to handle multiple chains of properties
	for _ = range a.Values {
		switch target := a.Parent.(type) {
		case *concepts.SimScalar:
			target.Reset()
		case *concepts.SimVector2:
			target.Reset()
		case *concepts.SimVector3:
			target.Reset()
		case *materials.Image:
			if a.Source.Name == "Source" {
				target.Load()
			}
		case *core.Expression:
			target.Construct(target.Code)
		}
	}
}

func (a *SetProperty) Act() {
	for _, v := range a.Values {
		origValue := reflect.ValueOf(v.Elem().Interface())
		a.Original = append(a.Original, origValue)
		v.Elem().Set(a.ToSet)
	}
	a.FireHooks()
	a.State().Modified = true
	a.ActionFinished(false)
}

func (a *SetProperty) Undo() {
	for i, v := range a.Values {
		fmt.Printf("Undo: %v\n", a.Original[i].String())
		v.Elem().Set(a.Original[i])
	}
	a.FireHooks()
}
func (a *SetProperty) Redo() {
	for _, v := range a.Values {
		fmt.Printf("Redo: %v\n", a.ToSet.String())
		v.Elem().Set(a.ToSet)
	}
	a.FireHooks()
}
