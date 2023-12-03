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

	"github.com/gotk3/gotk3/gdk"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
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
	}

	return nil
}

func (g *Grid) fieldEntityRef(index int, field *state.PropertyGridField) {
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
	rendText, _ := gtk.CellRendererTextNew()
	rendPix, _ := gtk.CellRendererPixbufNew()
	opts, _ := gtk.ListStoreNew(glib.TYPE_UINT64, glib.TYPE_STRING, glib.TYPE_OBJECT)
	box, _ := gtk.ComboBoxNewWithModel(opts)
	box.SetHExpand(true)
	box.PackStart(rendPix, true)
	box.PackStart(rendText, true)
	box.AddAttribute(rendPix, "pixbuf", 2)
	box.AddAttribute(rendText, "text", 1)

	for entity := range g.State().DB.EntityComponents {
		er := g.State().DB.EntityRef(entity)
		if archetypes.EntityRefIsMaterial(er) {
			listItem := opts.Append()
			pixbuf := g.pixbuf(er)
			name := strconv.FormatUint(er.Entity, 10)
			if named := concepts.NamedFromDb(er); named != nil {
				name = named.Name
			}
			opts.Set(listItem, []int{0, 1, 2}, []any{er.Entity, name, pixbuf})
			if !origValue.Nil() && er.Entity == origValue.Entity {
				box.SetActiveIter(listItem)
			}
		}
	}

	box.Connect("changed", func(_ *gtk.ComboBox) {
		selected, _ := box.GetActiveIter()
		value, _ := opts.GetValue(selected, 0)
		entity, _ := value.GoValue()
		er := g.State().DB.EntityRef(entity.(uint64))
		action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: reflect.ValueOf(er).Convert(field.Type.Elem())}
		g.NewAction(action)
		action.Act()
	})

	g.Container.Attach(box, 2, index, 2, 1)
}
