package properties

import (
	"reflect"

	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/gotk3/gotk3/gtk"
)

func (g *Grid) setFieldFile(entry *widget.Entry, field *state.PropertyGridField) {
	action := &actions.SetProperty{IEditor: g.IEditor, PropertyGridField: field, ToSet: reflect.ValueOf(entry.Text)}
	g.NewAction(action)
	action.Act()
}

func (g *Grid) fieldFile(field *state.PropertyGridField) {
	origValue := g.originalStrings(field)

	entry := widget.NewEntry()
	entry.SetText(origValue)
	entry.OnChanged = func(s string) {
		g.setFieldFile(entry, field)
	}

	button := widget.NewButtonWithIcon("...", theme.FolderOpenIcon(), func() {
		window, _ := g.Container.GetToplevel()
		native, _ := gtk.FileChooserNativeDialogNew("Load File...", window.(gtk.IWindow), gtk.FILE_CHOOSER_ACTION_OPEN, "_Load", "_Cancel")
		res := gtk.ResponseType(native.Run())

		if res == gtk.RESPONSE_ACCEPT {
			entry.SetText(native.GetFilename())
			g.setFieldFile(entry, field)
		}

	})
	g.FContainer.Add(container.NewBorder(nil, nil, nil, button, entry))
}
