package main

import (
	"reflect"

	"github.com/gotk3/gotk3/gtk"
)

type gridState struct {
	index   int
	visited map[interface{}]bool
}

type GridField struct {
	Name  string
	Value reflect.Value
	Index int
}

func (e *Editor) PropertyGridFieldString(state *gridState) {
	box, _ := gtk.EntryNew()
	box.SetHExpand(true)
	e.PropertyGrid.Attach(box, 2, state.index, 1, 1)
}

func (e *Editor) PropertyGridFields(obj interface{}, state *gridState) {
	v := reflect.ValueOf(obj)
	t := v.Type().Elem()
	if v.IsNil() || t.String() == "main.MapPoint" {
		return
	}

	label, _ := gtk.LabelNew(t.String())
	label.SetJustify(gtk.JUSTIFY_CENTER)
	e.PropertyGrid.Attach(label, 1, state.index, 2, 1)
	state.index++

	state.visited[obj] = true

	for i := 0; i < t.NumField(); i++ {
		field := t.FieldByIndex([]int{i})
		if display, ok := field.Tag.Lookup("editable"); ok {
			label, _ := gtk.LabelNew(display)
			label.SetJustify(gtk.JUSTIFY_LEFT)
			e.PropertyGrid.Attach(label, 1, state.index, 1, 1)
			e.PropertyGridFieldString(state)
			state.index++
		} else if field.Type.Kind() == reflect.Struct {
			child := v.Elem().Field(i).Addr().Interface()
			if !state.visited[child] {
				e.PropertyGridFields(child, state)
			}
		} else if field.Type.Kind() == reflect.Ptr {
			child := v.Elem().Field(i).Interface()
			if !state.visited[child] {
				e.PropertyGridFields(child, state)
			}
		}
	}
	e.PropertyGrid.ShowAll()
}

func (e *Editor) RefreshPropertyGrid() {
	e.PropertyGrid.GetChildren().Foreach(func(child interface{}) {
		e.PropertyGrid.Remove(child.(gtk.IWidget))
	})

	state := &gridState{visited: make(map[interface{}]bool), index: 1}
	for _, obj := range e.SelectedObjects {
		e.PropertyGridFields(obj, state)
	}
}
