// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"log"
	"reflect"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

type ChangeSliceMode int

const (
	AddSliceElementMode ChangeSliceMode = iota
	IncSliceElementMode
	DecSliceElementMode
	DeleteSliceElementMode
)

type ChangeSlice struct {
	state.Action
	Parent       any
	SlicePtr     reflect.Value
	Index        int
	ConcreteType reflect.Type
	Mode         ChangeSliceMode
}

func (a *ChangeSlice) Activate() {
	a.Redo()
	a.ActionFinished(false, true, false)
}

func (a *ChangeSlice) Undo() {
	// SlicePtr is something like: *[]<some type>
	oldSlice := a.SlicePtr.Elem()
	switch a.Mode {
	case AddSliceElementMode:
		if oldSlice.Len() > 0 {
			oldSlice.Slice(0, a.SlicePtr.Len()-1)
		}
	default:
		log.Printf("ChangeSlice.Undo: unimplemented: %v", a.Mode)
	}
	a.State().Universe.ActAllControllers(ecs.ControllerRecalculate)
}
func (a *ChangeSlice) Redo() {
	// SlicePtr is something like: *[]<some type>
	oldSlice := a.SlicePtr.Elem()
	switch a.Mode {
	case AddSliceElementMode:
		newSlice := reflect.Append(oldSlice, reflect.Zero(oldSlice.Type().Elem()))
		oldSlice.Set(newSlice)
		newValue := oldSlice.Index(oldSlice.Len() - 1)
		if newValue.Kind() == reflect.Pointer {
			// TODO: Use the a.ConcreteType field here?
			newValue.Set(reflect.New(newValue.Type().Elem()))
		}
		if universal, ok := newValue.Interface().(ecs.Universal); ok {
			universal.OnAttach(a.State().Universe)
		}

		if serializable, ok := newValue.Interface().(ecs.Serializable); ok {
			serializable.Construct(nil)
		} else if subSerializable, ok := newValue.Interface().(ecs.SubSerializable); ok {
			subSerializable.Construct(a.State().Universe, nil)
		}
	case DeleteSliceElementMode:

		newSlice := reflect.MakeSlice(oldSlice.Type(), oldSlice.Len()-1, oldSlice.Len()-1)
		// Copy elements
		j := 0
		for i := range oldSlice.Len() {
			// This approach means we don't have to validate the index
			if i == a.Index {
				continue
			}
			newSlice.Index(j).Set(oldSlice.Index(i))
			j++
		}
		oldSlice.Set(newSlice)
	case DecSliceElementMode:
		if a.Index <= 0 {
			return
		}
		v := oldSlice.Index(a.Index).Interface()
		prev := oldSlice.Index(a.Index - 1).Interface()
		oldSlice.Index(a.Index - 1).Set(reflect.ValueOf(v))
		oldSlice.Index(a.Index).Set(reflect.ValueOf(prev))
	case IncSliceElementMode:
		if a.Index >= oldSlice.Len()-1 {
			return
		}
		v := oldSlice.Index(a.Index).Interface()
		next := oldSlice.Index(a.Index + 1).Interface()
		oldSlice.Index(a.Index + 1).Set(reflect.ValueOf(v))
		oldSlice.Index(a.Index).Set(reflect.ValueOf(next))
	}
	a.State().Universe.ActAllControllers(ecs.ControllerRecalculate)
}
