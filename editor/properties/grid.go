package properties

import (
	"log"
	"reflect"
	"sort"
	"strings"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/components/core"

	"github.com/gotk3/gotk3/gtk"
)

type PropertyGridState struct {
	Fields           map[string]*state.PropertyGridField
	Visited          map[any]bool
	Depth            int
	ParentName       string
	ParentCollection *reflect.Value
	Parent           any
}

type Grid struct {
	state.IEditor
	Container *gtk.Grid
}

func (g *Grid) childFields(parentName string, childValue reflect.Value, state PropertyGridState, updateParent bool) {
	var child any
	if childValue.Type().Kind() == reflect.Struct {
		child = childValue.Addr().Interface()
	} else if childValue.Type().Kind() == reflect.Ptr || childValue.Type().Kind() == reflect.Interface {
		child = childValue.Interface()
	} else {
		log.Printf("%v, %v", childValue.String(), childValue.Type())
	}
	if !state.Visited[child] {
		state.ParentName = parentName
		if updateParent {
			state.Parent = child
		}
		g.gatherFields(child, state)
	}
}

func (g *Grid) gatherFields(obj any, pgs PropertyGridState) {
	objValue := reflect.ValueOf(obj)
	objType := objValue.Type().Elem()
	if objValue.IsNil() {
		return
	}

	pgs.Depth++
	pgs.Visited[obj] = true

	for i := 0; i < objType.NumField(); i++ {
		field := objType.FieldByIndex([]int{i})
		fieldValue := objValue.Elem().Field(i)
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
			/*if pgs.ParentName != "" {
				name = pgs.ParentName + "." + name
			}*/
			g.childFields(name, fieldValue, pgs, true)
		}
	}
}

func (g *Grid) Refresh(selection []any) {
	g.Container.GetChildren().Foreach(func(child any) {
		g.Container.Remove(child.(gtk.IWidget))
	})

	state := PropertyGridState{Visited: make(map[any]bool), Fields: make(map[string]*state.PropertyGridField)}
	for _, obj := range selection {
		switch target := obj.(type) {
		case *concepts.EntityRef:
			for _, c := range target.All() {
				if c == nil {
					continue
				}
				state.Parent = c
				n := strings.Split(reflect.TypeOf(c).String(), ".")
				state.ParentName = n[len(n)-1]
				g.gatherFields(c, state)
			}
		case *core.Segment:
			state.Parent = nil
			state.ParentName = "Segment"
			g.gatherFields(target, state)
		}
	}

	sorted := make([]string, len(state.Fields))
	i := 0
	for display := range state.Fields {
		sorted[i] = display
		i++
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	index := 1
	for _, display := range sorted {
		field := state.Fields[display]

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
		case *core.MaterialScale:
			g.fieldEnum(index, field, core.MaterialScaleValues())
		case *core.CollisionResponse:
			g.fieldEnum(index, field, core.CollisionResponseValues())
		case **concepts.EntityRef:
			g.fieldEntityRef(index, field)
		case *map[string]core.Sampleable:
			g.fieldMaterials(index, field)
		}
		index++
	}
	g.Container.ShowAll()
}
