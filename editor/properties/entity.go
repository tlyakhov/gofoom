// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"reflect"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"

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
	name := entity.Format()
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
		img.Image = g.IEditor.EntityImage(entity)
		img.SetMinSize(fyne.NewSquareSize(64))
		button.OnTapped = func() {
			g.SelectObjects(true, selection.SelectableFromEntity(entity))
		}
		fyne.Do(img.Refresh)
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

	entitySet := controllers.EntitiesByClass(editTypeTag)
	entities := make([]widget.TreeNodeID, 1)
	entities[0] = "0"

	for _, e := range entitySet {
		if e != 0 {
			entities = append(entities, e.Serialize())
		}
	}

	tree := widget.NewTree(func(tni widget.TreeNodeID) []widget.TreeNodeID {
		if tni != "" {
			return []string{}
		}
		return entities
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
		tree.Select(origValue.Serialize())
		title = editTypeTag + ": " + origValue.Format()
	}

	if field.Disabled() {
		tree.OnSelected = nil
	} else {
		tree.OnSelected = func(tni widget.TreeNodeID) {
			entity, _ := ecs.ParseEntity(tni)
			g.ApplySetPropertyAction(field, reflect.ValueOf(entity).Convert(field.Type.Elem()))
		}
	}

	c := container.New(&gridEntitySelectorLayout{Child: layout.NewStackLayout()}, tree)
	aiTree := widget.NewAccordionItem(concepts.TruncateString(title, 40), c)
	accordion := gridAddOrUpdateWidgetAtIndex[*widget.Accordion](g)
	accordion.Items = []*widget.AccordionItem{aiTree}
}

// This layout is just to make the selection list have a static size
type gridEntitySelectorLayout struct {
	Child fyne.Layout
}

func (erl *gridEntitySelectorLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	s := erl.Child.MinSize(objects)
	if s.Height < 500 {
		s.Height = 500
	}
	return s
}

func (erl *gridEntitySelectorLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	erl.Child.Layout(objects, containerSize)
}
