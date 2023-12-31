package main

import (
	"fmt"
	"reflect"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
)

func CreateMainMenu() {

	editor.ActionFileOpen.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierShortcutDefault}
	editor.ActionFileOpen.Menu = fyne.NewMenuItem("Open", func() {
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

	editor.ActionFileSaveAs.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift}
	editor.ActionFileSaveAs.Menu = fyne.NewMenuItem("Save As", func() {
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

	editor.ActionFileSave.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierShortcutDefault}
	editor.ActionFileSave.Menu = fyne.NewMenuItem("Save", func() {
		if editor.OpenFile == "" {
			editor.ActionFileSaveAs.Menu.Action()
			return
		}
		editor.DB.Save(editor.OpenFile)
		editor.Modified = false
		editor.UpdateTitle()
	})
	editor.ActionFileQuit.Menu = fyne.NewMenuItem("Quit", func() {})

	editor.ActionEditUndo.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyZ, Modifier: fyne.KeyModifierShortcutDefault}
	editor.ActionEditUndo.Menu = fyne.NewMenuItem("Undo", editor.UndoCurrent)
	editor.ActionEditRedo.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyY, Modifier: fyne.KeyModifierShortcutDefault}
	editor.ActionEditRedo.Menu = fyne.NewMenuItem("Redo", editor.RedoCurrent)

	editor.ActionEditDelete.NoModifier = true
	editor.ActionEditDelete.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyDelete, Modifier: 0}
	editor.ActionEditDelete.Menu = fyne.NewMenuItem("Delete", func() {
		action := &actions.Delete{IEditor: editor}
		editor.NewAction(action)
		action.Act()
	})

	editor.ActionEditToolSelect.NoModifier = true
	editor.ActionEditToolSelect.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyEscape}
	editor.ActionEditToolSelect.Menu = fyne.NewMenuItem("Select/Move", func() { editor.SwitchTool(state.ToolSelect) })
	editor.ActionEditSelectSegment.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyApostrophe, Modifier: fyne.KeyModifierShortcutDefault}
	editor.ActionEditSelectSegment.Menu = fyne.NewMenuItem("Select First/Next Segment", editor.ToolSelectSegment)
	editor.ActionEditRaiseCeil.NoModifier = true
	editor.ActionEditRaiseCeil.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyPageUp}
	editor.ActionEditRaiseCeil.Menu = fyne.NewMenuItem("Raise Selection Ceiling", func() { editor.MoveSurface(2, false, false) })
	editor.ActionEditLowerCeil.NoModifier = true
	editor.ActionEditLowerCeil.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyPageDown}
	editor.ActionEditLowerCeil.Menu = fyne.NewMenuItem("Lower Selection Ceiling", func() { editor.MoveSurface(-2, false, false) })
	editor.ActionEditRaiseFloor.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyPageUp, Modifier: fyne.KeyModifierShortcutDefault}
	editor.ActionEditRaiseFloor.Menu = fyne.NewMenuItem("Raise Selection Floor", func() { editor.MoveSurface(2, true, false) })
	editor.ActionEditLowerFloor.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyPageDown, Modifier: fyne.KeyModifierShortcutDefault}
	editor.ActionEditLowerFloor.Menu = fyne.NewMenuItem("Lower Selection Floor", func() { editor.MoveSurface(-2, true, false) })
	editor.ActionEditRaiseCeilSlope.NoModifier = true
	editor.ActionEditRaiseCeilSlope.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyRightBracket}
	editor.ActionEditRaiseCeilSlope.Menu = fyne.NewMenuItem("Raise Selection Ceiling Slope", func() { editor.MoveSurface(0.05, false, true) })
	editor.ActionEditLowerCeilSlope.NoModifier = true
	editor.ActionEditLowerCeilSlope.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyLeftBracket}
	editor.ActionEditLowerCeilSlope.Menu = fyne.NewMenuItem("Lower Selection Ceiling Slope", func() { editor.MoveSurface(-0.05, false, true) })
	editor.ActionEditRaiseFloorSlope.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyRightBracket, Modifier: fyne.KeyModifierShortcutDefault}
	editor.ActionEditRaiseFloorSlope.Menu = fyne.NewMenuItem("Raise Selection Floor Slope", func() { editor.MoveSurface(0.05, true, true) })
	editor.ActionEditLowerFloorSlope.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyLeftBracket, Modifier: fyne.KeyModifierShortcutDefault}
	editor.ActionEditLowerFloorSlope.Menu = fyne.NewMenuItem("Lower Selection Floor Slope", func() { editor.MoveSurface(-0.05, true, true) })
	editor.ActionEditRotateAnchor.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyR, Modifier: fyne.KeyModifierShortcutDefault}
	editor.ActionEditRotateAnchor.Menu = fyne.NewMenuItem("Rotate Slope Anchor", func() {
		action := &actions.RotateSegments{IEditor: editor}
		editor.NewAction(action)
		action.Act()
	})
	editor.ActionEditIncreaseGrid.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyUp, Modifier: fyne.KeyModifierShift}
	editor.ActionEditIncreaseGrid.Menu = fyne.NewMenuItem("Increase Grid Size", func() { editor.Current.Step *= 2 })
	editor.ActionEditDecreaseGrid.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyDown, Modifier: fyne.KeyModifierShift}
	editor.ActionEditDecreaseGrid.Menu = fyne.NewMenuItem("Decrease Grid Size", func() { editor.Current.Step /= 2 })

	menuFile := fyne.NewMenu("File", editor.ActionFileOpen.Menu, editor.ActionFileSave.Menu, editor.ActionFileSaveAs.Menu, editor.ActionFileQuit.Menu)
	menuEdit := fyne.NewMenu("Edit", editor.ActionEditUndo.Menu, editor.ActionEditRedo.Menu, fyne.NewMenuItemSeparator(),
		editor.ActionEditDelete.Menu, fyne.NewMenuItemSeparator(),
		editor.ActionEditToolSelect.Menu, fyne.NewMenuItemSeparator(),
		editor.ActionEditSelectSegment.Menu,
		editor.ActionEditRaiseCeil.Menu, editor.ActionEditLowerCeil.Menu, editor.ActionEditRaiseFloor.Menu, editor.ActionEditLowerFloor.Menu,
		editor.ActionEditRaiseCeilSlope.Menu, editor.ActionEditLowerCeilSlope.Menu, editor.ActionEditRaiseFloorSlope.Menu,
		editor.ActionEditLowerFloorSlope.Menu, editor.ActionEditRotateAnchor.Menu, editor.ActionEditIncreaseGrid.Menu, editor.ActionEditDecreaseGrid.Menu)
	mainMenu := fyne.NewMainMenu(menuFile, menuEdit)
	editor.Window.SetMainMenu(mainMenu)

	editor.MenuActions = make(map[string]*MenuAction)
	widgets := reflect.ValueOf(&editor.EditorWidgets).Elem()
	for i := 0; i < widgets.NumField(); i++ {
		f := widgets.Field(i)
		if f.Type() != reflect.TypeOf(MenuAction{}) {
			continue
		}
		action := f.Addr().Interface().(*MenuAction)
		editor.MenuActions[widgets.Type().Field(i).Name] = action
		if !action.NoModifier && action.Shortcut != nil {
			action.Menu.Shortcut = action.Shortcut
		}
	}
}
