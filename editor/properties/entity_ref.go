// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"image"
	"reflect"
	"strconv"

	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *Grid) imageForEntity(entity concepts.Entity) image.Image {
	w, h := 64, 64
	if img := materials.ImageFromDb(g.State().DB, entity); img != nil {
		return img.Image
	} else if text := materials.TextFromDb(g.State().DB, entity); text != nil && text.Rendered != nil {
		return text.Rendered.Image
	} else if solid := materials.SolidFromDb(g.State().DB, entity); solid != nil {
		img := image.NewNRGBA(image.Rect(0, 0, w, h))
		for i := 0; i < w*h; i++ {
			img.Pix[i*4+0] = uint8(solid.Diffuse.Now[0] * 255)
			img.Pix[i*4+1] = uint8(solid.Diffuse.Now[1] * 255)
			img.Pix[i*4+2] = uint8(solid.Diffuse.Now[2] * 255)
			img.Pix[i*4+3] = uint8(solid.Diffuse.Now[3] * 255)
		}
		return img
	} else if shader := materials.ShaderFromDb(g.State().DB, entity); shader != nil {
		if len(shader.Stages) == 0 {
			return nil
		}
		return g.imageForEntity(shader.Stages[0].Texture)
	}
	return image.NewRGBA(image.Rect(0, 0, w, h))
}

func (g *Grid) updateTreeNodeEntity(tni widget.TreeNodeID, b bool, co fyne.CanvasObject) {
	entity, _ := concepts.DeserializeEntity(tni)
	name := entity.NameString(g.State().DB)
	box := co.(*fyne.Container)
	img := box.Objects[0].(*canvas.Image)
	img.ScaleMode = canvas.ImageScaleSmooth
	img.FillMode = canvas.ImageFillContain
	img.Image = g.imageForEntity(entity)
	img.SetMinSize(fyne.NewSquareSize(64))
	label := box.Objects[1].(*widget.Label)
	label.SetText(name)
	button := box.Objects[2].(*widget.Button)
	button.OnTapped = func() {
		g.SelectObjects(true, core.SelectableFromEntityRef(g.State().DB, entity))
	}
}

func (g *Grid) fieldEntity(field *state.PropertyGridField) {
	// The value of this property is an Entity
	var origValue concepts.Entity
	if !field.Values[0].Elem().IsZero() {
		origValue = field.Values[0].Elem().Interface().(concepts.Entity)
	}

	editTypeTag, ok := field.Source.Tag.Lookup("edit_type")

	if !ok {
		return
	}

	// Create our combo box with pixbuf/string enum entries.
	refs := make([]widget.TreeNodeID, 0)

	for entity, c := range g.State().DB.EntityComponents {
		if c == nil {
			continue
		}
		if archetypes.EntityIsMaterial(g.State().DB, concepts.Entity(entity)) {
			refs = append(refs, strconv.Itoa(entity))
		}
	}
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
			widget.NewButtonWithIcon("", theme.LoginIcon(), nil),
		)
	}, g.updateTreeNodeEntity)
	title := "Select " + editTypeTag
	if origValue != 0 {
		tree.Select(origValue.Serialize())
		title = editTypeTag + ": " + origValue.NameString(g.State().DB)
	}
	tree.OnSelected = func(tni widget.TreeNodeID) {
		entity, _ := concepts.DeserializeEntity(tni)
		action := &actions.SetProperty{IEditor: g.IEditor,
			PropertyGridField: field,
			ToSet:             reflect.ValueOf(entity).Convert(field.Type.Elem()),
		}
		g.NewAction(action)
		action.Act()
	}
	c := container.New(&entityRefLayout{Child: layout.NewStackLayout()}, tree)
	aiTree := widget.NewAccordionItem(title, c)
	accordion := gridAddOrUpdateWidgetAtIndex[*widget.Accordion](g)
	accordion.Items = []*widget.AccordionItem{aiTree}
}

// This layout is just to make the selection list have a static size
type entityRefLayout struct {
	Child fyne.Layout
}

func (erl *entityRefLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	s := erl.Child.MinSize(objects)
	if s.Height < 200 {
		s.Height = 200
	}
	return s
}

func (erl *entityRefLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	erl.Child.Layout(objects, containerSize)
}
