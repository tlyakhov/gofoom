package properties

import (
	"fmt"
	"reflect"

	"github.com/tlyakhov/gofoom/editor/actions"

	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/materials"
	"github.com/tlyakhov/gofoom/texture"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func (g *Grid) pixbuf(obj interface{}) *gdk.Pixbuf {
	pbFromFile := func(filename string) *gdk.Pixbuf {
		pixbuf, err := gdk.PixbufNewFromFileAtSize(filename, 16, 16)
		if err != nil {
			fmt.Printf("Warning: %v", err)
			return nil
		}
		return pixbuf
	}
	switch target := obj.(type) {
	case *materials.PainfulLitSampled:
		return pbFromFile(target.Sampler.(*texture.Image).Source)
	case *materials.LitSampled:
		return pbFromFile(target.Sampler.(*texture.Image).Source)
	case *materials.Sky:
		return pbFromFile(target.Sampler.(*texture.Image).Source)
	}
	return nil
}

func (g *Grid) fieldSerializable(index int, field *pgField) {
	// The value of this property is a pointer to a type that implements ISerializable.
	var origValue string
	if !field.Values[0].Elem().IsNil() {
		origValue = field.Values[0].Elem().Interface().(concepts.ISerializable).GetBase().ID
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

	for id, mat := range g.State().World.Materials {
		listItem := opts.Append()
		pixbuf := g.pixbuf(mat)
		opts.Set(listItem, []int{0, 1}, []interface{}{id, pixbuf})
		if id == origValue {
			box.SetActiveIter(listItem)
		}
	}

	box.Connect("changed", func(_ *gtk.ComboBox) {
		selected, _ := box.GetActiveIter()
		value, _ := opts.GetValue(selected, 0)
		value2, _ := value.GoValue()
		ptr := g.State().World.Materials[value2.(string)]
		action := &actions.SetProperty{IEditor: g.IEditor, Fields: field.Values, ToSet: reflect.ValueOf(ptr).Convert(field.Type.Elem())}
		g.NewAction(action)
		action.Act()
	})

	g.Container.Attach(box, 2, index, 1, 1)
}
