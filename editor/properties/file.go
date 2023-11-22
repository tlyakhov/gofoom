package properties

import (
	"log"
	"reflect"

	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"github.com/gotk3/gotk3/gtk"
)

func (g *Grid) setFieldFile(entry *gtk.Entry, field *state.PropertyGridField) {
	origValue := g.originalStrings(field)
	text, err := entry.GetText()
	if err != nil {
		log.Printf("Couldn't get text from gtk.Entry. %v\n", err)
		entry.SetText(origValue)
		g.Container.GrabFocus()
		return
	}
	action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: reflect.ValueOf(text)}
	g.NewAction(action)
	action.Act()
	g.Container.GrabFocus()
}

func (g *Grid) fieldFile(index int, field *state.PropertyGridField) {
	origValue := g.originalStrings(field)

	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 4)

	entry, _ := gtk.EntryNew()
	entry.SetHExpand(true)
	entry.SetText(origValue)
	entry.Connect("activate", func(_ *gtk.Entry) {
		g.setFieldFile(entry, field)
	})
	box.PackStart(entry, true, true, 0)

	button, _ := gtk.ButtonNew()
	button.SetLabel("...")
	button.Connect("clicked", func(_ *gtk.Button) {
		window, _ := g.Container.GetToplevel()
		native, _ := gtk.FileChooserNativeDialogNew("Load File...", window.(gtk.IWindow), gtk.FILE_CHOOSER_ACTION_OPEN, "_Load", "_Cancel")
		res := gtk.ResponseType(native.Run())

		if res == gtk.RESPONSE_ACCEPT {
			entry.SetText(native.GetFilename())
			g.setFieldFile(entry, field)
		}

	})
	box.PackStart(button, false, false, 0)

	g.Container.Attach(box, 2, index, 2, 1)
}
