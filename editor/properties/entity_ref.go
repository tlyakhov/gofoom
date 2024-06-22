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

func (g *Grid) imageForRef(ref *concepts.EntityRef) image.Image {
	w, h := 64, 64
	if img := materials.ImageFromDb(ref); img != nil {
		return img.Image
	} else if text := materials.TextFromDb(ref); text != nil && text.Rendered != nil {
		return text.Rendered.Image
	} else if solid := materials.SolidFromDb(ref); solid != nil {
		img := image.NewNRGBA(image.Rect(0, 0, w, h))
		for i := 0; i < w*h; i++ {
			img.Pix[i*4+0] = uint8(solid.Diffuse.Now[0] * 255)
			img.Pix[i*4+1] = uint8(solid.Diffuse.Now[1] * 255)
			img.Pix[i*4+2] = uint8(solid.Diffuse.Now[2] * 255)
			img.Pix[i*4+3] = uint8(solid.Diffuse.Now[3] * 255)
		}
		return img
	} else if shader := materials.ShaderFromDb(ref); shader != nil {
		if len(shader.Stages) == 0 {
			return nil
		}
		return g.imageForRef(shader.Stages[0].Texture)
	}
	return image.NewRGBA(image.Rect(0, 0, w, h))
}

func (g *Grid) updateTreeNodeEntityRef(tni widget.TreeNodeID, b bool, co fyne.CanvasObject) {
	entity, _ := strconv.ParseUint(tni, 10, 64)
	ref := g.State().DB.EntityRef(entity)
	name := ref.NameString()
	box := co.(*fyne.Container)
	img := box.Objects[0].(*canvas.Image)
	img.ScaleMode = canvas.ImageScaleSmooth
	img.FillMode = canvas.ImageFillContain
	img.Image = g.imageForRef(ref)
	img.SetMinSize(fyne.NewSquareSize(64))
	label := box.Objects[1].(*widget.Label)
	label.SetText(name)
	button := box.Objects[2].(*widget.Button)
	button.OnTapped = func() {
		g.IEditor.SelectObject(core.SelectableFromEntityRef(ref), true)
	}
}

func (g *Grid) fieldEntityRef(field *state.PropertyGridField) {
	// The value of this property is an EntityRef or pointer to EntityRef
	var origValue *concepts.EntityRef
	if !field.Values[0].Elem().IsNil() {
		origValue = field.Values[0].Elem().Interface().(*concepts.EntityRef)
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
		er := g.State().DB.EntityRef(uint64(entity))
		if archetypes.EntityRefIsMaterial(er) {
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
		)
	}, g.updateTreeNodeEntityRef)
	title := "Select " + editTypeTag
	if !origValue.Nil() {
		tree.Select(strconv.FormatUint(origValue.Entity, 10))
		title = editTypeTag + ": " + origValue.NameString()
	}
	tree.OnSelected = func(tni widget.TreeNodeID) {
		entity, _ := strconv.ParseUint(tni, 10, 64)
		ref := g.State().DB.EntityRef(entity)
		action := &actions.SetProperty{IEditor: g.IEditor,
			PropertyGridField: field,
			ToSet:             reflect.ValueOf(ref).Convert(field.Type.Elem()),
		}
		g.NewAction(action)
		action.Act()
	}
	c := container.New(&entityRefLayout{Child: layout.NewStackLayout()}, tree)
	aitem := widget.NewAccordionItem(title, c)
	accordion := gridAddOrUpdateWidgetAtIndex[*widget.Accordion](g)
	accordion.Items = []*widget.AccordionItem{aitem}
}
