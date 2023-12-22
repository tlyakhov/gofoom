package properties

import (
	"image/color"
	"log"
	"reflect"
	"strconv"

	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/gotk3/gotk3/gdk"
)

func (g *Grid) pixbufFromFile(filename string) *gdk.Pixbuf {
	pixbuf, err := gdk.PixbufNewFromFileAtSize(filename, 16, 16)
	if err != nil {
		log.Printf("Warning: %v", err)
		return nil
	}
	return pixbuf
}

func (g *Grid) pixbufFromColor(c color.NRGBA) *gdk.Pixbuf {
	pixbuf, err := gdk.PixbufNew(gdk.COLORSPACE_RGB, false, 8, 16, 16)
	if err != nil {
		log.Printf("Warning: %v", err)
		return nil
	}
	pixels := pixbuf.GetPixels()
	for i := 0; i < len(pixels)-2; i += 3 {
		pixels[i] = c.R
		pixels[i+1] = c.G
		pixels[i+2] = c.B
	}
	return pixbuf
}

func (g *Grid) pixbuf(er *concepts.EntityRef) *gdk.Pixbuf {
	if img := materials.ImageFromDb(er); img != nil {
		return g.pixbufFromFile(img.Source)
	} else if solid := materials.SolidFromDb(er); solid != nil {
		return g.pixbufFromColor(solid.Diffuse)
	} else if shader := materials.ShaderFromDb(er); shader != nil {
		if len(shader.Stages) == 0 {
			return nil
		}
		return g.pixbuf(shader.Stages[0].Texture)
	}

	return nil
}

func (g *Grid) fieldEntityRef(field *state.PropertyGridField) {
	// The value of this property is an EntityRef or pointer to EntityRef
	/*var origValue *concepts.EntityRef
	if !field.Values[0].Elem().IsNil() {
		origValue = field.Values[0].Elem().Interface().(*concepts.EntityRef)
	}*/

	_, ok := field.Source.Tag.Lookup("edit_type")

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
			return nil
		}
		return refs
	}, func(tni widget.TreeNodeID) bool {
		return false
	}, func(b bool) fyne.CanvasObject {
		return container.NewHBox(canvas.NewRasterFromImage(nil), widget.NewLabel("Template"))
	}, func(tni widget.TreeNodeID, b bool, co fyne.CanvasObject) {
		entity, _ := strconv.ParseUint(tni, 10, 64)
		ref := g.State().DB.EntityRef(entity)
		name := string(tni)
		if named := concepts.NamedFromDb(ref); named != nil {
			name = named.Name
		}
		box := co.(*fyne.Container)
		//		raster := box.Objects[0].(*canvas.Raster)
		label := box.Objects[1].(*widget.Label)
		label.SetText(name)
	})
	aitem := widget.NewAccordionItem("Test", tree)
	accordion := widget.NewAccordion(aitem)
	g.FContainer.Add(accordion)
}

func (g *Grid) fieldEntityRef2(field *state.PropertyGridField) {
	// The value of this property is an EntityRef or pointer to EntityRef
	var origValue *concepts.EntityRef
	if !field.Values[0].Elem().IsNil() {
		origValue = field.Values[0].Elem().Interface().(*concepts.EntityRef)
	}

	_, ok := field.Source.Tag.Lookup("edit_type")

	if !ok {
		return
	}

	// Create our combo box with pixbuf/string enum entries.
	opts := make([]string, 0)
	optsValues := make([]uint64, 0)
	selected := 0

	for entity, c := range g.State().DB.EntityComponents {
		if c == nil {
			continue
		}
		er := g.State().DB.EntityRef(uint64(entity))
		if archetypes.EntityRefIsMaterial(er) {
			//pixbuf := g.pixbuf(er)
			name := strconv.FormatUint(er.Entity, 10)
			if named := concepts.NamedFromDb(er); named != nil {
				name = named.Name
			}
			opts = append(opts, name)
			optsValues = append(optsValues, er.Entity)
			if !origValue.Nil() && er.Entity == origValue.Entity {
				selected = len(opts) - 1
			}
		}
	}

	selectEntry := widget.NewSelect(opts, nil)
	selectEntry.OnChanged = func(selected string) {
		er := g.State().DB.EntityRef(optsValues[selectEntry.SelectedIndex()])
		action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: reflect.ValueOf(er).Convert(field.Type.Elem())}
		g.NewAction(action)
		action.Act()
	}
	selectEntry.SetSelectedIndex(selected)

	button := widget.NewButtonWithIcon("", theme.MoreHorizontalIcon(), func() {
		er := g.State().DB.EntityRef(optsValues[selectEntry.SelectedIndex()])
		g.IEditor.SelectObjects([]any{er}, true)
	})
	g.FContainer.Add(container.NewBorder(nil, nil, nil, button, selectEntry))
}
