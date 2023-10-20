package properties

import (
	"fmt"
	"reflect"
	"sort"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/core"

	"github.com/gotk3/gotk3/gtk"
)

type pgState struct {
	Fields           map[string]*state.PropertyGridField
	Visited          map[interface{}]bool
	Depth            int
	ParentName       string
	ParentCollection *reflect.Value
	Parent           interface{}
}

type Grid struct {
	state.IEditor
	Container *gtk.Grid
}

func (g *Grid) childFields(parentName string, childValue reflect.Value, state pgState, updateParent bool) {
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
		if updateParent {
			state.Parent = child
		}
		g.gatherFields(child, state)
	}
}

func (g *Grid) gatherFields(obj interface{}, pgs pgState) {
	v := reflect.ValueOf(obj)
	t := v.Type().Elem()
	if v.IsNil() || t.String() == "main.MapPoint" {
		return
	}

	pgs.Depth++
	pgs.Visited[obj] = true

	for i := 0; i < t.NumField(); i++ {
		field := t.FieldByIndex([]int{i})
		fieldValue := v.Elem().Field(i)
		tag, ok := field.Tag.Lookup("editable")
		if !ok {
			continue
		}
		display := tag
		if pgs.ParentName != "" {
			display = pgs.ParentName + "." + display
		}

		if tag == "^" {
			// Include the child fields as part of the parent.
			// This is nice for embedded Golang structs so we don't have
			// a giant nested hierarchy.
			g.childFields(pgs.ParentName, fieldValue, pgs, false)
			continue
		}

		gf, ok := pgs.Fields[display]
		if !ok {
			gf = &state.PropertyGridField{
				Name:             display,
				Depth:            pgs.Depth,
				Type:             fieldValue.Addr().Type(),
				Source:           &field,
				ParentName:       pgs.ParentName,
				ParentCollection: pgs.ParentCollection,
				Unique:           make(map[string]reflect.Value),
				Parent:           pgs.Parent,
			}
			pgs.Fields[display] = gf
		}

		gf.Values = append(gf.Values, fieldValue.Addr())
		gf.Unique[fieldValue.String()] = fieldValue.Addr()

		if field.Type.Kind() == reflect.Map {
			keys := fieldValue.MapKeys()
			for _, key := range keys {
				name := field.Name + "[" + key.String() + "]"
				if pgs.ParentName != "" {
					name = pgs.ParentName + "." + name
				}
				pgs2 := pgs
				pgs2.ParentCollection = &fieldValue
				g.childFields(name, fieldValue.MapIndex(key), pgs2, true)
			}
		} else if field.Type.Name() == "SimScalar" || field.Type.Name() == "SimVector2" || field.Type.Name() == "SimVector3" {
			delete(pgs.Fields, display)
			name := display
			if pgs.ParentName != "" {
				name = pgs.ParentName + "." + name
			}
			g.childFields(name, fieldValue, pgs, false)
		}
	}
}

func (g *Grid) Refresh(selection []concepts.ISerializable) {
	g.Container.GetChildren().Foreach(func(child interface{}) {
		g.Container.Remove(child.(gtk.IWidget))
	})

	state := pgState{Visited: make(map[interface{}]bool), Fields: make(map[string]*state.PropertyGridField)}
	for _, obj := range selection {
		state.Parent = obj
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
			label.SetJustify(gtk.JUSTIFY_FILL)
			label.SetHExpand(true)
			label.SetHAlign(gtk.ALIGN_CENTER)
			g.Container.Attach(label, 1, index, 2, 1)
			lastParentName = field.ParentName
			index++
		}

		if !field.Values[0].CanInterface() {
			continue
		}
		label, _ := gtk.LabelNew(field.Short())
		label.SetTooltipText(field.Name)
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
		case *map[string]core.Sampleable:
			g.fieldMaterials(index, field)
		}
		index++
	}
	g.Container.ShowAll()
}
