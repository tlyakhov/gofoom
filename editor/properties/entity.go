// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"reflect"
	"strconv"

	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/selection"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *Grid) updateTreeNodeEntity(editTypeTag string, tni widget.TreeNodeID, _ bool, co fyne.CanvasObject) {
	entity, _ := ecs.ParseEntity(tni)
	name := entity.NameString(g.State().ECS)
	box := co.(*fyne.Container)
	img := box.Objects[0].(*canvas.Image)
	img.Hidden = entity == 0
	label := box.Objects[1].(*widget.Label)
	button := box.Objects[2].(*widget.Button)
	button.Hidden = entity == 0

	label.SetText(name)

	if entity != 0 {
		img.ScaleMode = canvas.ImageScaleSmooth
		img.FillMode = canvas.ImageFillContain
		if editTypeTag == "Sector" || editTypeTag == "Material" {
			// TODO: Make images for other entity types like actions
			img.Image = g.IEditor.EntityImage(entity, editTypeTag != "Material")
		}
		img.SetMinSize(fyne.NewSquareSize(64))
		button.OnTapped = func() {
			g.SelectObjects(true, selection.SelectableFromEntity(g.State().ECS, entity))
		}
	}
}

func (g *Grid) fieldEntity(field *state.PropertyGridField) {
	// The value of this property is an Entity
	var origValue ecs.Entity
	if !field.Values[0].Deref().IsZero() {
		origValue = field.Values[0].Interface().(ecs.Entity)
	}

	editTypeTag, ok := field.Source.Tag.Lookup("edit_type")

	if !ok {
		return
	}

	// Create our combo box with pixbuf/string enum entries.
	refs := make([]widget.TreeNodeID, 1)
	refs[0] = "0"

	// TODO: Optimize this by using columns directly. This is unwieldy
	g.State().ECS.Entities.Range(func(entity uint32) {
		if editTypeTag == "Material" && archetypes.EntityIsMaterial(g.State().ECS, ecs.Entity(entity)) {
			refs = append(refs, strconv.Itoa(int(entity)))
		} else if editTypeTag == "Sector" && core.GetSector(g.State().ECS, ecs.Entity(entity)) != nil {
			refs = append(refs, strconv.Itoa(int(entity)))
		} else if editTypeTag == "Action" && behaviors.GetActionWaypoint(g.State().ECS, ecs.Entity(entity)) != nil {
			refs = append(refs, strconv.Itoa(int(entity)))
		} else if editTypeTag == "Weapon" && behaviors.GetWeaponClass(g.State().ECS, ecs.Entity(entity)) != nil {
			refs = append(refs, strconv.Itoa(int(entity)))
		}
	})
	tree := widget.NewTree(func(tni widget.TreeNodeID) []widget.TreeNodeID {
		if tni != "" {
			return []string{}
		}
		return refs
	}, func(tni widget.TreeNodeID) bool {
		return tni == ""
	}, func(b bool) fyne.CanvasObject {
		return container.NewHBox(
			canvas.NewImageFromImage(nil),
			widget.NewLabel("Template"),
			widget.NewButtonWithIcon("", theme.MoreHorizontalIcon(), nil),
		)
	}, func(tni widget.TreeNodeID, b bool, co fyne.CanvasObject) {
		g.updateTreeNodeEntity(editTypeTag, tni, b, co)
	})
	title := "Select " + editTypeTag
	if origValue != 0 {
		tree.Select(origValue.String())
		title = editTypeTag + ": " + origValue.NameString(g.State().ECS)
	}
	tree.OnSelected = func(tni widget.TreeNodeID) {
		entity, _ := ecs.ParseEntity(tni)
		g.ApplySetPropertyAction(field, reflect.ValueOf(entity).Convert(field.Type.Elem()))
	}
	c := container.New(&gridEntitySelectorLayout{Child: layout.NewStackLayout()}, tree)
	aiTree := widget.NewAccordionItem(title, c)
	accordion := gridAddOrUpdateWidgetAtIndex[*widget.Accordion](g)
	accordion.Items = []*widget.AccordionItem{aiTree}
}

// This layout is just to make the selection list have a static size
type gridEntitySelectorLayout struct {
	Child fyne.Layout
}

func (erl *gridEntitySelectorLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	s := erl.Child.MinSize(objects)
	if s.Height < 200 {
		s.Height = 200
	}
	return s
}

func (erl *gridEntitySelectorLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	erl.Child.Layout(objects, containerSize)
}
