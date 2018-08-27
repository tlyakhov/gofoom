package main

import (
	"reflect"
	"sort"

	"github.com/tlyakhov/gofoom/concepts"

	"github.com/tlyakhov/gofoom/core"

	"github.com/gotk3/gotk3/gtk"
)

type GridField struct {
	Name       string
	Values     []reflect.Value
	Type       reflect.Type
	ParentName string
	Depth      int
	Source     *reflect.StructField
}

type GridState struct {
	Fields  map[string]*GridField
	Visited map[interface{}]bool
	Depth   int
}

func (e *Editor) PropertyGridFields(obj interface{}, state GridState) {
	v := reflect.ValueOf(obj)
	t := v.Type().Elem()
	if v.IsNil() || t.String() == "main.MapPoint" {
		return
	}

	state.Depth++
	state.Visited[obj] = true

	for i := 0; i < t.NumField(); i++ {
		field := t.FieldByIndex([]int{i})
		display, ok := field.Tag.Lookup("editable")
		if !ok {
			continue
		}

		if display != "^" {
			gf, ok := state.Fields[display]
			if !ok {
				gf = &GridField{
					Name:   display,
					Depth:  state.Depth,
					Type:   v.Elem().Field(i).Addr().Type(),
					Source: &field,
				}
				state.Fields[display] = gf
			}
			gf.Values = append(gf.Values, v.Elem().Field(i).Addr())

			continue
		}

		var child interface{}
		if field.Type.Kind() == reflect.Struct {
			child = v.Elem().Field(i).Addr().Interface()
		} else if field.Type.Kind() == reflect.Ptr {
			child = v.Elem().Field(i).Interface()
		}
		if !state.Visited[child] {
			e.PropertyGridFields(child, state)
		}
	}
}

func (e *Editor) RefreshPropertyGrid() {
	e.PropertyGrid.GetChildren().Foreach(func(child interface{}) {
		e.PropertyGrid.Remove(child.(gtk.IWidget))
	})

	state := GridState{Visited: make(map[interface{}]bool), Fields: make(map[string]*GridField)}
	for _, obj := range e.SelectedObjects {
		e.PropertyGridFields(obj, state)
	}

	sorted := make([]string, len(state.Fields))
	i := 0
	for display := range state.Fields {
		sorted[i] = display
		i++
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		f1 := state.Fields[sorted[i]]
		f2 := state.Fields[sorted[j]]
		if f1.Depth > f2.Depth {
			return true
		} else if f1.Depth < f2.Depth {
			return false
		}
		return f1.Name < f2.Name
	})

	var lastParentName string
	index := 1
	for _, display := range sorted {
		field := state.Fields[display]
		if field.ParentName != lastParentName {
			label, _ := gtk.LabelNew(field.ParentName)
			label.SetJustify(gtk.JUSTIFY_CENTER)
			e.PropertyGrid.Attach(label, 1, index, 2, 1)
			lastParentName = field.ParentName
			index++
		}

		label, _ := gtk.LabelNew(field.Name)
		label.SetJustify(gtk.JUSTIFY_LEFT)
		label.SetHExpand(true)
		label.SetHAlign(gtk.ALIGN_START)
		e.PropertyGrid.Attach(label, 1, index, 1, 1)
		if !field.Values[0].CanInterface() {
			continue
		}
		switch field.Values[0].Interface().(type) {
		case *bool:
			e.PropertyGridFieldBool(index, field)
		case *string:
			e.PropertyGridFieldString(index, field)
		case *float64:
			e.PropertyGridFieldFloat64(index, field)
		case *concepts.Vector2:
			e.PropertyGridFieldVector2(index, field)
		case *concepts.Vector3:
			e.PropertyGridFieldVector3(index, field)
		case *core.MaterialBehavior:
			e.PropertyGridFieldEnum(index, field, core.MaterialBehaviorValues())
		case *core.CollisionResponse:
			e.PropertyGridFieldEnum(index, field, core.CollisionResponseValues())
		case *concepts.ISerializable:
			e.PropertyGridFieldCollection(index, field)
		}
		index++
	}
	e.PropertyGrid.ShowAll()
}
