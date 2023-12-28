package properties

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
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
	entry.OnSubmitted = func(s string) {
		g.setFieldFile(entry, field)
	}

	button := widget.NewButtonWithIcon("...", theme.FolderOpenIcon(), func() {
		dlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			if err != nil {
				g.Alert(fmt.Sprintf("Error loading file: %v", err))
				return
			}
			if uc == nil {
				return
			}
			entry.SetText(uc.URI().Path())
			g.setFieldFile(entry, field)
		}, g.GridWindow)
		g.SetDialogLocation(dlg, entry.Text)

		if entry.Text != "" {
			dlg.SetFileName(entry.Text)
			absPath, err := filepath.Abs(entry.Text)
			if err != nil {
				log.Printf("Load file: error making absolute path from %v", entry.Text)
				absPath, _ = os.Getwd()
			}
			dir := filepath.Dir(absPath)
			uri := storage.NewFileURI(dir)
			lister, err := storage.ListerForURI(uri)
			if err != nil {
				log.Printf("Load file: error making lister from %v", dir)
			} else {
				dlg.SetLocation(lister)
			}
		}
		dlg.Resize(fyne.NewSize(1000, 700))
		dlg.SetConfirmText("Load file")
		dlg.SetDismissText("Cancel")
		dlg.Show()
	})
	g.FContainer.Add(container.NewBorder(nil, nil, nil, button, entry))
}
