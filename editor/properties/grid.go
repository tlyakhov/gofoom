package properties

import (
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/components/core"

	"github.com/gotk3/gotk3/glib"
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
		g.fieldsFromObject(child, state)
	}
}

func (g *Grid) fieldsFromObject(obj any, pgs PropertyGridState) {
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
			if editTypeTag, ok := field.Tag.Lookup("edit_type"); ok {
				gf.EditType = editTypeTag
			}
		}

		gf.Values = append(gf.Values, fieldValue.Addr())
		gf.Unique[fieldValue.String()] = fieldValue.Addr()

		if field.Type.Kind() == reflect.Map && false {
			keys := fieldValue.MapKeys()
			for _, key := range keys {
				name := field.Name + "[" + key.String() + "]"
				pgsChild := pgs
				pgsChild.ParentCollection = &fieldValue
				g.childFields(name, fieldValue.MapIndex(key), pgsChild, true)
			}
		} else if field.Type.Name() == "SimScalar" || field.Type.Name() == "SimVector2" || field.Type.Name() == "SimVector3" || field.Type.Name() == "Expression" {
			delete(pgs.Fields, display)
			name := display
			g.childFields(name, fieldValue, pgs, true)
		} else if field.Type.Kind() == reflect.Slice {
			for i := 0; i < fieldValue.Len(); i++ {
				name := fmt.Sprintf("%v[%v]", display, i)
				pgsChild := pgs
				pgsChild.ParentCollection = &fieldValue
				g.childFields(name, fieldValue.Index(i), pgsChild, true)
			}
		}
	}
}

func (g *Grid) fieldsFromSelection(selection []any) *PropertyGridState {
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
				g.fieldsFromObject(c, state)
			}
		case *core.Segment:
			state.Parent = nil
			state.ParentName = "Segment"
			g.fieldsFromObject(target, state)
		}
	}
	return &state
}

func (g *Grid) AddEntityControls(selection []any) {
	index := 1
	entities := make([]uint64, 0)
	entityList := ""
	componentList := make([]bool, len(concepts.DbTypes().Indexes))
	for _, obj := range selection {
		if len(entityList) > 0 {
			entityList += ", "
		}
		switch target := obj.(type) {
		case *concepts.EntityRef:
			entities = append(entities, target.Entity)
			entityList += strconv.FormatUint(target.Entity, 10)
			for index, c := range target.All() {
				componentList[index] = (c != nil)
			}
		case *core.Segment:
			entities = append(entities, target.Sector.Entity)
			entityList += strconv.FormatUint(target.Sector.Entity, 10)
			for index, c := range target.Sector.All() {
				componentList[index] = (c != nil)
			}
		}
	}

	label, _ := gtk.LabelNew("")
	label.SetMarkup(fmt.Sprintf("<b>Entity [%v]</b>", concepts.TruncateString(entityList, 10)))
	label.SetTooltipText(entityList)
	label.SetJustify(gtk.JUSTIFY_LEFT)
	label.SetHExpand(true)
	label.SetHAlign(gtk.ALIGN_START)
	g.Container.Attach(label, 1, index, 1, 1)

	rendText, _ := gtk.CellRendererTextNew()
	opts, _ := gtk.ListStoreNew(glib.TYPE_INT, glib.TYPE_STRING)
	box, _ := gtk.ComboBoxNewWithModel(opts)
	box.SetHExpand(true)
	box.PackStart(rendText, true)
	box.AddAttribute(rendText, "text", 1)
	for index, t := range concepts.DbTypes().Types {
		if t == nil || componentList[index] || index == concepts.AttachedComponentIndex {
			continue
		}
		listItem := opts.Append()
		opts.Set(listItem, []int{0, 1}, []any{index, t.String()})
		if t.String() == "behaviors.Proximity" {
			box.SetActiveIter(listItem)
		}
	}
	g.Container.Attach(box, 2, index, 1, 1)

	button, _ := gtk.ButtonNew()
	button.SetHExpand(true)
	button.SetLabel("Add")
	button.Connect("clicked", func(_ *gtk.Button) {
		selected, _ := box.GetActiveIter()
		value, _ := opts.GetValue(selected, 0)
		componentIndex, _ := value.GoValue()

		action := &actions.AddComponent{IEditor: g.IEditor, Index: componentIndex.(int), Entities: entities}
		g.NewAction(action)
		action.Act()
		g.Container.GrabFocus()
	})
	g.Container.Attach(button, 3, index, 1, 1)
}

func (g *Grid) Refresh(selection []any) {
	g.Container.GetChildren().Foreach(func(child any) {
		g.Container.Remove(child.(gtk.IWidget))
	})

	state := g.fieldsFromSelection(selection)

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
	if len(selection) > 0 {
		g.AddEntityControls(selection)
		index++
	}
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

		if field.EditType == "Component" {
			g.fieldComponent(index, field)
			index++
			continue
		}

		switch field.Values[0].Interface().(type) {
		case *bool:
			g.fieldBool(index, field)
		case *string:
			if field.EditType == "file" {
				g.fieldFile(index, field)
			} else {
				g.fieldString(index, field)
			}
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
		case *[]core.Trigger:
			g.fieldSlice(index, field)
		}
		index++
	}
	g.Container.ShowAll()
}
