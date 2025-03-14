// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package properties

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *Grid) fieldFile(field *state.PropertyGridField) {
	origValue := g.originalStrings(field)

	entry := widget.NewEntry()
	entry.SetText(origValue)
	entry.OnSubmitted = func(s string) {
		g.ApplySetPropertyAction(field, reflect.ValueOf(entry.Text))
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
			wd, _ := os.Getwd()
			target := uc.URI().Path()
			if target, err = filepath.Rel(wd, target); err != nil {
				g.Alert(fmt.Sprintf("Error loading file: %v", err))
				return
			}
			entry.SetText(target)
			g.ApplySetPropertyAction(field, reflect.ValueOf(entry.Text))
		}, g.GridWindow)

		if entry.Text != "" {
			g.SetDialogLocation(dlg, entry.Text)

			/*			dlg.SetFileName(filepath.Base(entry.Text))
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
						}*/
		}
		dlg.Resize(fyne.NewSize(1000, 700))
		dlg.SetConfirmText("Load file")
		dlg.SetDismissText("Cancel")
		dlg.Show()
	})

	if field.Disabled() {
		button.Disable()
		entry.Disable()
	} else {
		button.Enable()
		entry.Enable()
	}

	c := gridAddOrUpdateWidgetAtIndex[*fyne.Container](g)
	c.Layout = layout.NewBorderLayout(nil, nil, nil, button)
	c.Objects = []fyne.CanvasObject{entry, button}
}
