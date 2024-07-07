// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"image"
	"math"
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

func (g *Grid) entityBorderColor(entity concepts.Entity) *concepts.Vector4 {
	if materials.ShaderFromDb(g.State().DB, entity) != nil {
		return &concepts.Vector4{1.0, 0.0, 1.0, 0.5}
	}
	return &concepts.Vector4{0.0, 0.0, 0.0, 0.0}
}

// TODO: We should cache these
func (g *Grid) imageForEntity(entity concepts.Entity) image.Image {
	w, h := 64, 64
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	buffer := img.Pix
	border := g.entityBorderColor(entity)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			g.MaterialSampler.ScreenX = x * g.MaterialSampler.ScreenWidth / w
			g.MaterialSampler.ScreenY = y * g.MaterialSampler.ScreenHeight / h
			g.MaterialSampler.Angle = float64(x) * math.Pi * 2.0 / float64(w)
			c := g.MaterialSampler.SampleShader(entity, nil, float64(x)/float64(w), float64(y)/float64(h), 1.0)
			if x <= 1 || y <= 1 || x >= w-2 || y >= h-2 {
				c.AddPreMulColorSelf(border)
			}
			index := x*4 + y*img.Stride
			buffer[index+0] = uint8(concepts.Clamp(c[0]*255, 0, 255))
			buffer[index+1] = uint8(concepts.Clamp(c[1]*255, 0, 255))
			buffer[index+2] = uint8(concepts.Clamp(c[2]*255, 0, 255))
			buffer[index+3] = uint8(concepts.Clamp(c[3]*255, 0, 255))
		}
	}
	return img
}

func (g *Grid) updateTreeNodeEntity(tni widget.TreeNodeID, b bool, co fyne.CanvasObject) {
	entity, _ := concepts.ParseEntity(tni)
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
		g.SelectObjects(true, core.SelectableFromEntity(g.State().DB, entity))
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
		tree.Select(origValue.Format())
		title = editTypeTag + ": " + origValue.NameString(g.State().DB)
	}
	tree.OnSelected = func(tni widget.TreeNodeID) {
		entity, _ := concepts.ParseEntity(tni)
		action := &actions.SetProperty{IEditor: g.IEditor,
			PropertyGridField: field,
			ToSet:             reflect.ValueOf(entity).Convert(field.Type.Elem()),
		}
		g.NewAction(action)
		action.Act()
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
