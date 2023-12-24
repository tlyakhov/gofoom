package properties

import (
	"reflect"

	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
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
		dlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			entry.SetText(uc.URI().Path())
			g.setFieldFile(entry, field)
		}, g.GridWindow)
		dlg.SetFileName(entry.Text)
		dlg.SetConfirmText("Load file")
		dlg.SetDismissText("Cancel")
		dlg.Show()
	})
	g.FContainer.Add(container.NewBorder(nil, nil, nil, button, entry))
}
