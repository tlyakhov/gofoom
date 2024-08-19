// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	rs "tlyakhov/gofoom/render/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type PropertyGridState struct {
	Fields           map[string]*state.PropertyGridField
	Visited          concepts.Set[any]
	Depth            int
	ParentName       string
	ParentCollection *reflect.Value
	Ancestors        []any
	ParentField      *state.PropertyGridField
	Entity           ecs.Entity
}

type Grid struct {
	state.IEditor
	GridWidget      *fyne.Container
	GridWindow      fyne.Window
	MaterialSampler rs.MaterialSampler

	refreshIndex int
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
	if !state.Visited.Contains(child) {
		state.ParentName = parentName
		if updateParent {
			ancestors := make([]any, len(state.Ancestors)+1)
			copy(ancestors, state.Ancestors)
			ancestors[len(ancestors)-1] = child
			state.Ancestors = ancestors
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
	pgs.Visited.Add(obj)

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
				Name:       display,
				Depth:      pgs.Depth,
				Type:       fieldValue.Addr().Type(),
				Sort:       100,
				Source:     &field,
				ParentName: pgs.ParentName,
				Parent:     pgs.ParentField,
				Unique:     make(map[string]reflect.Value),
			}
			pgs.Fields[display] = gf
			if editTypeTag, ok := field.Tag.Lookup("edit_type"); ok {
				gf.EditType = editTypeTag
			}
			if editSortTag, ok := field.Tag.Lookup("edit_sort"); ok {
				gf.Sort, _ = strconv.Atoi(editSortTag)
			}
		}

		gf.Values = append(gf.Values, &state.PropertyGridFieldValue{
			Entity:           pgs.Entity,
			Value:            fieldValue.Addr(),
			ParentCollection: pgs.ParentCollection,
			Ancestors:        pgs.Ancestors,
		})
		gf.Unique[fieldValue.String()] = fieldValue.Addr()

		if gf.IsEmbeddedType() {
			// Animations are a special case
			if !gf.Type.Elem().AssignableTo(reflect.TypeFor[ecs.Animated]()) {
				delete(pgs.Fields, display)
			}
			name := display
			g.childFields(name, fieldValue, pgs, true)
		} else if field.Type.Kind() == reflect.Slice {
			for i := 0; i < fieldValue.Len(); i++ {
				name := fmt.Sprintf("%v[%v]", display, i)
				pgsChild := pgs
				pgsChild.ParentField = gf
				pgsChild.ParentCollection = &fieldValue
				g.childFields(name, fieldValue.Index(i), pgsChild, true)
			}
		}
	}
}

func (g *Grid) fieldsFromSelection(selection *core.Selection) *PropertyGridState {
	pgs := PropertyGridState{Visited: make(concepts.Set[any]), Fields: make(map[string]*state.PropertyGridField)}
	for _, s := range selection.Exact {
		switch s.Type {
		case core.SelectableHi:
			fallthrough
		case core.SelectableLow:
			fallthrough
		case core.SelectableMid:
			fallthrough
		case core.SelectableSectorSegment:
			pgs.Ancestors = []any{s.SectorSegment}
			pgs.ParentName = "Segment"
			pgs.Entity = s.Entity
			g.fieldsFromObject(s.SectorSegment, pgs)
			continue
		}

		if s.Entity == 0 {
			continue
		}
		for _, c := range s.ECS.AllComponents(s.Entity) {
			if c == nil {
				continue
			}
			pgs.Ancestors = []any{c}
			n := strings.Split(reflect.TypeOf(c).String(), ".")
			pgs.ParentName = n[len(n)-1]
			pgs.Entity = s.Entity
			g.fieldsFromObject(c, pgs)
		}
	}
	return &pgs
}

// Confusing syntax. The constraint ensures that our underlying type has pointer
// receiver methods that implement fyne.CanvasObject
func gridAddOrUpdateWidgetAtIndex[PT interface {
	*T
	fyne.CanvasObject
}, T any](g *Grid) PT {
	var ptr PT = new(T)
	return gridAddOrUpdateAtIndex(g, ptr)
}

// Confusing syntax. The constraint ensures that our underlying type has pointer
// receiver methods that implement fyne.CanvasObject
func gridAddOrUpdateAtIndex[PT interface {
	*T
	fyne.CanvasObject
}, T any](g *Grid, newInstance PT) PT {
	if g.refreshIndex < len(g.GridWidget.Objects) {
		if element, ok := g.GridWidget.Objects[g.refreshIndex].(PT); ok {
			g.refreshIndex++
			return element
		}
		g.GridWidget.Objects[g.refreshIndex] = newInstance
		g.refreshIndex++
		return newInstance
	}

	g.GridWidget.Objects = append(g.GridWidget.Objects, newInstance)
	g.refreshIndex++
	return newInstance
}

func (g *Grid) AddEntityControls(selection *core.Selection) {
	entities := make([]ecs.Entity, 0)
	entityList := ""
	componentList := make([]bool, len(ecs.Types().Indexes))
	for _, s := range selection.Exact {
		if len(entityList) > 0 {
			entityList += ", "
		}

		switch s.Type {
		case core.SelectableHi:
			fallthrough
		case core.SelectableLow:
			fallthrough
		case core.SelectableMid:
			fallthrough
		case core.SelectableSectorSegment:
			label := gridAddOrUpdateWidgetAtIndex[*widget.Label](g)
			label.Text = fmt.Sprintf("Sector [%v]", s.Sector.Entity)
			label.TextStyle.Bold = true
			//label.SetTooltipText(entityList)
			label.Alignment = fyne.TextAlignLeading

			button := gridAddOrUpdateWidgetAtIndex[*widget.Button](g)
			button.SetText("Select parent sector")
			button.SetIcon(theme.LoginIcon())
			button.OnTapped = func() {
				g.SelectObjects(true, core.SelectableFromEntity(g.State().ECS, s.Sector.Entity))
			}
		}
		entities = append(entities, s.Entity)
		entityList += s.Entity.Format()
		for index, c := range s.ECS.AllComponents(s.Entity) {
			componentList[index] = (c != nil)
		}
	}

	if len(entityList) == 0 {
		return
	}

	label := gridAddOrUpdateWidgetAtIndex[*widget.Label](g)
	label.Text = fmt.Sprintf("Entity [%v]", concepts.TruncateString(entityList, 10))
	label.TextStyle.Bold = true
	//label.SetTooltipText(entityList)
	label.Alignment = fyne.TextAlignLeading

	opts := make([]string, 0)
	optsIndices := make([]int, 0)
	for index, t := range ecs.Types().Types {
		if t == nil || componentList[index] || index == ecs.AttachedComponentIndex {
			continue
		}
		opts = append(opts, t.String())
		optsIndices = append(optsIndices, index)
	}
	selectComponent := widget.NewSelect(opts, func(s string) {})

	button := widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), func() {
		optsIndex := selectComponent.SelectedIndex()
		if optsIndex < 0 {
			return
		}
		action := &actions.AddComponent{IEditor: g.IEditor, Index: optsIndices[optsIndex], Entities: entities}
		g.NewAction(action)
		action.Act()

	})
	c := gridAddOrUpdateWidgetAtIndex[*fyne.Container](g)
	c.Layout = layout.NewBorderLayout(nil, nil, nil, button)
	c.Objects = []fyne.CanvasObject{selectComponent, button}
	c.Refresh()
}

func (g *Grid) sortedFields(state *PropertyGridState) []string {
	sorted := make([]string, len(state.Fields))
	i := 0
	for display := range state.Fields {
		sorted[i] = display
		i++
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	return sorted
}

func (g *Grid) Refresh(selection *core.Selection) {
	g.refreshIndex = 0

	/*	distinctTypes := make(map[core.SelectableType]bool)
		for _, s := range selection.Exact {
			distinctTypes[s.Type] = true
		}

		if len(distinctTypes) > 2 {
			g.GridWidget.Objects = make([]fyne.CanvasObject, 0)
			return
		}*/

	if len(selection.Exact) > 0 {
		g.AddEntityControls(selection)
	}

	state := g.fieldsFromSelection(selection)
	for _, display := range g.sortedFields(state) {
		field := state.Fields[display]

		if !field.Values[0].Value.CanInterface() {
			continue
		}
		label := gridAddOrUpdateWidgetAtIndex[*widget.Label](g)
		label.Text = field.Short()
		//label.SetTooltipText(field.Name)
		label.Alignment = fyne.TextAlignLeading

		if field.EditType == "Component" {
			label.Importance = widget.HighImportance
			label.TextStyle.Bold = true
			g.fieldComponent(field)
			continue
		}
		label.Importance = widget.MediumImportance
		label.TextStyle.Bold = false
		//label.Wrapping = fyne.TextWrapWord

		switch field.Values[0].Value.Interface().(type) {
		case *bool:
			g.fieldBool(field)
		case *string:
			switch field.EditType {
			case "file":
				g.fieldFile(field)
			case "multi-line-string":
				g.fieldString(field, true)
			default:
				g.fieldString(field, false)
			}
		case *concepts.Set[string]:
			g.fieldString(field, false)
		case *float32:
			g.fieldNumber(field)
		case *float64:
			g.fieldNumber(field)
		case *int:
			g.fieldNumber(field)
		case *uint32:
			g.fieldNumber(field)
		case *uint64:
			g.fieldNumber(field)
		case *concepts.Vector2:
			fieldStringLikeType[concepts.Vector2](g, field)
		case *concepts.Vector3:
			fieldStringLikeType[concepts.Vector3](g, field)
		case *concepts.Vector4:
			fieldStringLikeType[concepts.Vector4](g, field)
		case **concepts.Vector2:
			fieldStringLikeType[*concepts.Vector2](g, field)
		case **concepts.Vector3:
			fieldStringLikeType[*concepts.Vector3](g, field)
		case **concepts.Vector4:
			fieldStringLikeType[*concepts.Vector4](g, field)
		case *concepts.Matrix2:
			g.fieldMatrix2(field)
		case *core.CollisionResponse:
			g.fieldEnum(field, core.CollisionResponseValues())
		case *core.BodyShadow:
			g.fieldEnum(field, core.BodyShadowValues())
		case *core.ScriptStyle:
			g.fieldEnum(field, core.ScriptStyleValues())
		case *ecs.AnimationLifetime:
			g.fieldEnum(field, ecs.AnimationLifetimeValues())
		case *ecs.AnimationCoordinates:
			g.fieldEnum(field, ecs.AnimationCoordinatesValues())
		case *materials.ShaderFlags:
			g.fieldEnum(field, materials.ShaderFlagsValues())
		case *ecs.Entity:
			g.fieldEntity(field)
		case *[]*core.Script:
			g.fieldSlice(field)
		case *[]*materials.Sprite:
			g.fieldSlice(field)
		case *[]*materials.ShaderStage:
			g.fieldSlice(field)
		case *[]*behaviors.InventorySlot:
			g.fieldSlice(field)
		case *[]ecs.Animated:
			g.fieldSlice(field)
		case *concepts.TweeningFunc:
			g.fieldTweeningFunc(field)
		case **ecs.Animation[int]:
			g.fieldAnimation(field)
		case **ecs.Animation[float64]:
			g.fieldAnimation(field)
		case **ecs.Animation[concepts.Vector2]:
			g.fieldAnimation(field)
		case **ecs.Animation[concepts.Vector3]:
			g.fieldAnimation(field)
		case **ecs.Animation[concepts.Vector4]:
			g.fieldAnimation(field)
		case **ecs.DynamicValue[int]:
			g.fieldAnimationTarget(field)
		case **ecs.DynamicValue[float64]:
			g.fieldAnimationTarget(field)
		case **ecs.DynamicValue[concepts.Vector2]:
			g.fieldAnimationTarget(field)
		case **ecs.DynamicValue[concepts.Vector3]:
			g.fieldAnimationTarget(field)
		case **ecs.DynamicValue[concepts.Vector4]:
			g.fieldAnimationTarget(field)
		default:
			g.refreshIndex++
			g.GridWidget.Add(widget.NewLabel("Unavailable"))
		}
	}
	if len(g.GridWidget.Objects) > g.refreshIndex {
		g.GridWidget.Objects = g.GridWidget.Objects[:g.refreshIndex]
	}
	g.GridWidget.Refresh()
}

func (g *Grid) Focus(o fyne.CanvasObject) {
	/*
		if c := fyne.CurrentApp().Driver().CanvasForObject(o); c != nil {
			c.Focus(o.(fyne.Focusable))
		}
	*/
}

func (g *Grid) ApplySetPropertyAction(field *state.PropertyGridField, v reflect.Value) {
	action := &actions.SetProperty{
		IEditor:           g.IEditor,
		PropertyGridField: field,
	}
	action.AssignAll(v)
	g.NewAction(action)
	action.Act()
}
