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
	"tlyakhov/gofoom/components/materials"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type PropertyGridState struct {
	Fields           map[string]*state.PropertyGridField
	Visited          map[any]bool
	Depth            int
	ParentName       string
	ParentCollection *reflect.Value
	Parent           any
	Ref              *concepts.EntityRef
}

type Grid struct {
	state.IEditor
	GridWidget *fyne.Container
	GridWindow fyne.Window
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
				Ref:              pgs.Ref,
			}
			pgs.Fields[display] = gf
			if editTypeTag, ok := field.Tag.Lookup("edit_type"); ok {
				gf.EditType = editTypeTag
			}
		}

		gf.Values = append(gf.Values, fieldValue.Addr())
		gf.Unique[fieldValue.String()] = fieldValue.Addr()

		if gf.IsEmbeddedType() {
			// Animations are a special case
			if !gf.Type.Elem().AssignableTo(concepts.ReflectType[concepts.Animated]()) {
				delete(pgs.Fields, display)
			}
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
				state.Ref = target
				g.fieldsFromObject(c, state)
			}
		case *core.Segment:
			state.Parent = nil
			state.ParentName = "Segment"
			state.Ref = target.Sector.Ref()
			g.fieldsFromObject(target, state)
		}
	}
	return &state
}

func (g *Grid) AddEntityControls(selection []any) {
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
	label := widget.NewLabel(fmt.Sprintf("Entity [%v]", concepts.TruncateString(entityList, 10)))
	label.TextStyle.Bold = true
	//label.SetTooltipText(entityList)
	label.Alignment = fyne.TextAlignLeading
	g.GridWidget.Objects = append(g.GridWidget.Objects, label)

	opts := make([]string, 0)
	optsIndices := make([]int, 0)
	for index, t := range concepts.DbTypes().Types {
		if t == nil || componentList[index] || index == concepts.AttachedComponentIndex {
			continue
		}
		opts = append(opts, t.String())
		optsIndices = append(optsIndices, index)
	}
	selectComponent := widget.NewSelect(opts, func(s string) {
	})

	button := widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), func() {
		optsIndex := selectComponent.SelectedIndex()
		if optsIndex < 0 {
			return
		}
		action := &actions.AddComponent{IEditor: g.IEditor, Index: optsIndices[optsIndex], Entities: entities}
		g.NewAction(action)
		action.Act()

	})
	g.GridWidget.Objects = append(g.GridWidget.Objects, container.NewBorder(nil, nil, nil, button, selectComponent))
}

func (g *Grid) Refresh(selection []any) {
	g.GridWidget.Objects = nil

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
	if len(selection) > 0 {
		g.AddEntityControls(selection)
	}
	for _, display := range sorted {
		field := state.Fields[display]

		if !field.Values[0].CanInterface() {
			continue
		}
		label := widget.NewLabel(field.Short())
		//label.SetTooltipText(field.Name)
		label.Alignment = fyne.TextAlignLeading
		g.GridWidget.Objects = append(g.GridWidget.Objects, label)

		if field.EditType == "Component" {
			label.Importance = widget.HighImportance
			label.TextStyle.Bold = true
			g.fieldComponent(field)
			continue
		}

		switch field.Values[0].Interface().(type) {
		case *bool:
			g.fieldBool(field)
		case *string:
			if field.EditType == "file" {
				g.fieldFile(field)
			} else {
				g.fieldString(field)
			}
		case *float64:
			g.fieldNumber(field)
		case *int:
			g.fieldNumber(field)
		case *concepts.Vector2:
			fieldStringLikeType[*concepts.Vector2](g, field)
		case *concepts.Vector3:
			fieldStringLikeType[*concepts.Vector3](g, field)
		case *concepts.Vector4:
			fieldStringLikeType[*concepts.Vector4](g, field)
		case *concepts.Matrix2:
			fieldStringLikeType[*concepts.Matrix2](g, field)
		case *materials.SurfaceScale:
			g.fieldEnum(field, materials.SurfaceScaleValues())
		case *core.CollisionResponse:
			g.fieldEnum(field, core.CollisionResponseValues())
		case *core.BodyShadow:
			g.fieldEnum(field, core.BodyShadowValues())
		case *core.ScriptStyle:
			g.fieldEnum(field, core.ScriptStyleValues())
		case *concepts.AnimationLifetime:
			g.fieldEnum(field, concepts.AnimationLifetimeValues())
		case *concepts.AnimationCoordinates:
			g.fieldEnum(field, concepts.AnimationCoordinatesValues())
		case **concepts.EntityRef:
			g.fieldEntityRef(field)
		case *[]*core.Script:
			g.fieldSlice(field)
		case *[]*materials.Sprite:
			g.fieldSlice(field)
		case *[]*materials.ShaderStage:
			g.fieldSlice(field)
		case *[]concepts.Animated:
			g.fieldSlice(field)
		case **concepts.Animation[int]:
			g.fieldAnimation(field)
		case **concepts.Animation[float64]:
			g.fieldAnimation(field)
		case **concepts.Animation[concepts.Vector2]:
			g.fieldAnimation(field)
		case **concepts.Animation[concepts.Vector3]:
			g.fieldAnimation(field)
		case **concepts.Animation[concepts.Vector4]:
			g.fieldAnimation(field)
		case **concepts.SimVariable[int]:
			g.fieldAnimationTarget(field)
		case **concepts.SimVariable[float64]:
			g.fieldAnimationTarget(field)
		case **concepts.SimVariable[concepts.Vector2]:
			g.fieldAnimationTarget(field)
		case **concepts.SimVariable[concepts.Vector3]:
			g.fieldAnimationTarget(field)
		case **concepts.SimVariable[concepts.Vector4]:
			g.fieldAnimationTarget(field)
		default:
			g.GridWidget.Add(widget.NewLabel("Unavailable"))
		}
	}
}

func (g *Grid) Focus(o fyne.CanvasObject) {
	/*
		if c := fyne.CurrentApp().Driver().CanvasForObject(o); c != nil {
			c.Focus(o.(fyne.Focusable))
		}
	*/
}
