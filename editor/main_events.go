package main

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func MainOpen(obj *glib.Object) {
	native, _ := gtk.FileChooserNativeDialogNew("Open world...", &editor.Window.Window, gtk.FILE_CHOOSER_ACTION_OPEN, "_Open", "_Cancel")
	res := gtk.ResponseType(native.Run())

	if res == gtk.RESPONSE_ACCEPT {
		editor.Load(native.GetFilename())
	}
}

func MainSave(obj *glib.Object) {
	if editor.OpenFile == "" {
		MainSaveAs(obj)
		return
	}
	editor.DB.Save(editor.OpenFile)
	editor.Modified = false
	editor.UpdateTitle()
}

func MainSaveAs(obj *glib.Object) {
	native, _ := gtk.FileChooserNativeDialogNew("Save world...", &editor.Window.Window, gtk.FILE_CHOOSER_ACTION_SAVE, "_Save", "_Cancel")
	native.SetDoOverwriteConfirmation(true)

	if editor.OpenFile != "" {
		native.SetCurrentName(editor.OpenFile)
	} else {
		native.SetCurrentName("untitled.json")
	}

	res := gtk.ResponseType(native.Run())
	if res == gtk.RESPONSE_ACCEPT {
		editor.DB.Save(native.GetFilename())
		editor.OpenFile = native.GetFilename()
		editor.Modified = false
		editor.UpdateTitle()
	}
}
