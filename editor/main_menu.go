package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func CreateMainMenu() {

	menuFileOpen := fyne.NewMenuItem("Open", func() {
		dlg := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			if err != nil {
				editor.Alert(fmt.Sprintf("Error loading world: %v", err))
				return
			}
			if uc == nil {
				return
			}
			editor.Load(uc.URI().Path())
		}, editor.Window)
		editor.SetDialogLocation(dlg, editor.OpenFile)
		dlg.Resize(fyne.NewSize(1000, 700))
		dlg.SetConfirmText("Load world")
		dlg.SetDismissText("Cancel")
		dlg.Show()
	})
	menuFileSaveAs := fyne.NewMenuItem("Save As", func() {
		dlg := dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
			if err != nil {
				editor.Alert(fmt.Sprintf("Error saving world: %v", err))
				return
			}
			if uc == nil {
				return
			}
			editor.DB.Save(uc.URI().Path())
			editor.OpenFile = uc.URI().Path()
			editor.Modified = false
			editor.UpdateTitle()
		}, editor.Window)
		target := editor.OpenFile
		if target == "" {
			target = "untitled.json"
		}
		editor.SetDialogLocation(dlg, target)

		dlg.Resize(fyne.NewSize(1000, 700))
		dlg.SetConfirmText("Save world")
		dlg.SetDismissText("Cancel")
		dlg.Show()
	})
	menuFileSave := fyne.NewMenuItem("Save", func() {
		if editor.OpenFile == "" {
			menuFileSaveAs.Action()
			return
		}
		editor.DB.Save(editor.OpenFile)
		editor.Modified = false
		editor.UpdateTitle()
	})
	menuFileQuit := fyne.NewMenuItem("Quit", func() {})
	menuFile := fyne.NewMenu("File", menuFileOpen, menuFileSave, menuFileSaveAs, menuFileQuit)
	menuEdit := fyne.NewMenu("Edit")
	mainMenu := fyne.NewMainMenu(menuFile, menuEdit)
	editor.Window.SetMainMenu(mainMenu)
}
