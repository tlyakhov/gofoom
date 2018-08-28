package properties

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/editor/state"

	"github.com/tlyakhov/gofoom/core"

	"github.com/gotk3/gotk3/gtk"
)

type pgField struct {
	Name       string
	Values     []reflect.Value
	Unique     map[string]reflect.Value
	Type       reflect.Type
	ParentName string
	Depth      int
	Source     *reflect.StructField
}

type pgState struct {
	Fields     map[string]*pgField
	Visited    map[interface{}]bool
	Depth      int
	ParentName string
}

type Grid struct {
	state.IEditor
	Container *gtk.Grid
}

func (g *Grid) childFields(parentName string, childValue reflect.Value, state pgState) {
	var child interface{}
	if childValue.Type().Kind() == reflect.Struct {
		child = childValue.Addr().Interface()
	} else if childValue.Type().Kind() == reflect.Ptr || childValue.Type().Kind() == reflect.Interface {
		child = childValue.Interface()
	} else {
		fmt.Printf("%v, %v", childValue.String(), childValue.Type())
	}
	if !state.Visited[child] {
		state.ParentName = parentName
		g.gatherFields(child, state)
	}
}

func (g *Grid) gatherFields(obj interface{}, state pgState) {
	v := reflect.ValueOf(obj)
	t := v.Type().Elem()
	if v.IsNil() || t.String() == "main.MapPoint" {
		return
	}

	fmt.Printf("%v\n", t)
	state.Depth++
	state.Visited[obj] = true

	for i := 0; i < t.NumField(); i++ {
		field := t.FieldByIndex([]int{i})
		fieldValue := v.Elem().Field(i)
		display, ok := field.Tag.Lookup("editable")
		if !ok {
			continue
		}
		fmt.Println(display)

		if field.Type.Kind() == reflect.Map {
			keys := fieldValue.MapKeys()
			for _, key := range keys {
				name := field.Name + "[" + key.String() + "]"
				if state.ParentName != "" {
					name = state.ParentName + "." + name
				}
				g.childFields(name, fieldValue.MapIndex(key), state)
			}
		} else if display != "^" {
			if state.ParentName != "" {
				display = state.ParentName + "." + display
			}
			gf, ok := state.Fields[display]
			if !ok {
				gf = &pgField{
					Name:   display,
					Depth:  state.Depth,
					Type:   fieldValue.Addr().Type(),
					Source: &field,
					Unique: make(map[string]reflect.Value),
				}
				state.Fields[display] = gf
			}

			gf.Values = append(gf.Values, fieldValue.Addr())
			gf.Unique[fieldValue.String()] = fieldValue.Addr()

			continue
		} else {
			name := field.Name
			if state.ParentName != "" {
				name = state.ParentName + "." + name
			}
			g.childFields(name, fieldValue, state)
		}
	}
}

func (g *Grid) Refresh(selection []concepts.ISerializable) {
	g.Container.GetChildren().Foreach(func(child interface{}) {
		g.Container.Remove(child.(gtk.IWidget))
	})

	state := pgState{Visited: make(map[interface{}]bool), Fields: make(map[string]*pgField)}
	for _, obj := range selection {
		g.gatherFields(obj, state)
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
		/*if f1.Depth > f2.Depth {
			return true
		} else if f1.Depth < f2.Depth {
			return false
		}*/
		return f1.Name < f2.Name
	})

	var lastParentName string
	index := 1
	for _, display := range sorted {
		field := state.Fields[display]
		if field.ParentName != lastParentName {
			label, _ := gtk.LabelNew(field.ParentName)
			label.SetJustify(gtk.JUSTIFY_CENTER)
			g.Container.Attach(label, 1, index, 2, 1)
			lastParentName = field.ParentName
			index++
		}

		if !field.Values[0].CanInterface() {
			continue
		}
		label, _ := gtk.LabelNew(field.Name)
		label.SetJustify(gtk.JUSTIFY_LEFT)
		label.SetHExpand(true)
		label.SetHAlign(gtk.ALIGN_START)
		g.Container.Attach(label, 1, index, 1, 1)

		switch field.Values[0].Interface().(type) {
		case *bool:
			g.fieldBool(index, field)
		case *string:
			g.fieldString(index, field)
		case *float64:
			g.fieldFloat64(index, field)
		case *concepts.Vector2:
			g.fieldVector2(index, field)
		case *concepts.Vector3:
			g.fieldVector3(index, field)
		case *core.MaterialBehavior:
			g.fieldEnum(index, field, core.MaterialBehaviorValues())
		case *core.CollisionResponse:
			g.fieldEnum(index, field, core.CollisionResponseValues())
		case *concepts.ISerializable:
			g.fieldSerializable(index, field)
		}
		index++
	}
	g.Container.ShowAll()
}
