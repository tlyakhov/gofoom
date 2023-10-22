package properties

import (
	"fmt"
	"image/color"
	"reflect"

	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/materials"
	"tlyakhov/gofoom/texture"

	"github.com/gotk3/gotk3/gdk"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func (g *Grid) pixbufSampler(sampler texture.ISampler) *gdk.Pixbuf {
	pbFromFile := func(filename string) *gdk.Pixbuf {
		pixbuf, err := gdk.PixbufNewFromFileAtSize(filename, 16, 16)
		if err != nil {
			fmt.Printf("Warning: %v", err)
			return nil
		}
		return pixbuf
	}
	pbFromColor := func(c color.NRGBA) *gdk.Pixbuf {
		pixbuf, err := gdk.PixbufNew(gdk.COLORSPACE_RGB, false, 8, 16, 16)
		if err != nil {
			fmt.Printf("Warning: %v", err)
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
	switch target := sampler.(type) {
	case *texture.Image:
		return pbFromFile(target.Source)
	case *texture.Solid:
		return pbFromColor(target.Diffuse)
	}
	return nil
}

func (g *Grid) pixbuf(obj interface{}) *gdk.Pixbuf {

	switch target := obj.(type) {
	case *materials.PainfulLitSampled:
		return g.pixbufSampler(target.Sampler)
	case *materials.LitSampled:
		return g.pixbufSampler(target.Sampler)
	case *materials.Sky:
		return g.pixbufSampler(target.Sampler)
	}
	return nil
}

func (g *Grid) fieldSerializable(index int, field *state.PropertyGridField) {
	// The value of this property is a pointer to a type that implements ISerializable.
	var origValue string
	if !field.Values[0].Elem().IsNil() {
		origValue = field.Values[0].Elem().Interface().(concepts.ISerializable).GetBase().Name
	}

	_, ok := field.Source.Tag.Lookup("edit_type")

	if !ok {
		return
	}

	// Create our combo box with pixbuf/string enum entries.
	rendText, _ := gtk.CellRendererTextNew()
	rendPix, _ := gtk.CellRendererPixbufNew()
	opts, _ := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_OBJECT)
	box, _ := gtk.ComboBoxNewWithModel(opts)
	box.SetHExpand(true)
	box.PackStart(rendPix, true)
	box.PackStart(rendText, true)
	box.AddAttribute(rendPix, "pixbuf", 1)
	box.AddAttribute(rendText, "text", 0)

	for name, mat := range g.State().World.Materials {
		listItem := opts.Append()
		pixbuf := g.pixbuf(mat)
		opts.Set(listItem, []int{0, 1}, []interface{}{name, pixbuf})
		if name == origValue {
			box.SetActiveIter(listItem)
		}
	}

	box.Connect("changed", func(_ *gtk.ComboBox) {
		selected, _ := box.GetActiveIter()
		value, _ := opts.GetValue(selected, 0)
		value2, _ := value.GoValue()
		ptr := g.State().World.Materials[value2.(string)]
		action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: reflect.ValueOf(ptr).Convert(field.Type.Elem())}
		g.NewAction(action)
		action.Act()
	})

	g.Container.Attach(box, 2, index, 1, 1)
}
