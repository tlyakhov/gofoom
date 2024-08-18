// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"reflect"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
)

type MenuAction struct {
	Shortcut   *desktop.CustomShortcut
	Menu       *fyne.MenuItem
	NoModifier bool
}

type EditorMenu struct {
	FileOpen   MenuAction
	FileSaveAs MenuAction
	FileSave   MenuAction
	FileQuit   MenuAction

	EditUndo   MenuAction
	EditRedo   MenuAction
	EditDelete MenuAction
	EditCut    MenuAction
	EditCopy   MenuAction
	EditPaste  MenuAction

	EditSelectSegment   MenuAction
	EditRaiseCeil       MenuAction
	EditLowerCeil       MenuAction
	EditRaiseFloor      MenuAction
	EditLowerFloor      MenuAction
	EditRaiseCeilSlope  MenuAction
	EditLowerCeilSlope  MenuAction
	EditRaiseFloorSlope MenuAction
	EditLowerFloorSlope MenuAction
	EditRotateAnchor    MenuAction
	EditIncreaseGrid    MenuAction
	EditDecreaseGrid    MenuAction

	EditTweakSurfaceLeft    MenuAction
	EditTweakSurfaceRight   MenuAction
	EditTweakSurfaceUp      MenuAction
	EditTweakSurfaceDown    MenuAction
	EditTweakSurfaceCW      MenuAction
	EditTweakSurfaceCCW     MenuAction
	EditTweakSurfaceLarger  MenuAction
	EditTweakSurfaceSmaller MenuAction

	ToolsSelect             MenuAction
	ToolsAddBody            MenuAction
	ToolsAddSector          MenuAction
	ToolsAddInternalSegment MenuAction
	ToolsSplitSegment       MenuAction
	ToolsSplitSector        MenuAction
	ToolsAlignGrid          MenuAction
	ToolsNewShader          MenuAction

	ViewSectorEntities MenuAction

	BehaviorsPause MenuAction
	BehaviorsReset MenuAction

	MenuActions map[string]*MenuAction
}

func CreateMainMenu() {

	editor.FileOpen.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierShortcutDefault}
	editor.FileOpen.Menu = fyne.NewMenuItem("Open", func() {
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

	editor.FileSaveAs.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift}
	editor.FileSaveAs.Menu = fyne.NewMenuItem("Save As", func() {
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

	editor.FileSave.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierShortcutDefault}
	editor.FileSave.Menu = fyne.NewMenuItem("Save", func() {
		if editor.OpenFile == "" {
			editor.FileSaveAs.Menu.Action()
			return
		}
		editor.DB.Save(editor.OpenFile)
		editor.Modified = false
		editor.UpdateTitle()
	})
	editor.FileQuit.Menu = fyne.NewMenuItem("Quit", func() {})

	editor.EditUndo.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyZ, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditUndo.Menu = fyne.NewMenuItem("Undo", editor.UndoCurrent)
	editor.EditRedo.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyY, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditRedo.Menu = fyne.NewMenuItem("Redo", editor.RedoCurrent)

	editor.EditDelete.NoModifier = true
	editor.EditDelete.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyDelete, Modifier: 0}
	editor.EditDelete.Menu = fyne.NewMenuItem("Delete", func() {
		action := &actions.Delete{IEditor: editor}
		editor.NewAction(action)
		action.Act()
	})

	editor.EditCut.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyX, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditCut.Menu = fyne.NewMenuItem("Cut", func() {
	})

	editor.EditCopy.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyC, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditCopy.Menu = fyne.NewMenuItem("Copy", func() {
		action := &actions.Copy{IEditor: editor}
		editor.NewAction(action)
		action.Act()
	})

	editor.EditPaste.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyV, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditPaste.Menu = fyne.NewMenuItem("Paste", func() {
		action := &actions.Paste{Transform: actions.Transform{IEditor: editor}}
		editor.NewAction(action)
		action.Act()
	})

	editor.EditSelectSegment.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyApostrophe, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditSelectSegment.Menu = fyne.NewMenuItem("Select First/Next Segment", editor.ToolSelectSegment)
	editor.EditRaiseCeil.NoModifier = true
	editor.EditRaiseCeil.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyPageUp}
	editor.EditRaiseCeil.Menu = fyne.NewMenuItem("Raise Selection Ceiling", func() { editor.MoveSurface(2, false, false) })
	editor.EditLowerCeil.NoModifier = true
	editor.EditLowerCeil.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyPageDown}
	editor.EditLowerCeil.Menu = fyne.NewMenuItem("Lower Selection Ceiling", func() { editor.MoveSurface(-2, false, false) })
	editor.EditRaiseFloor.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyPageUp, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditRaiseFloor.Menu = fyne.NewMenuItem("Raise Selection Floor", func() { editor.MoveSurface(2, true, false) })
	editor.EditLowerFloor.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyPageDown, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditLowerFloor.Menu = fyne.NewMenuItem("Lower Selection Floor", func() { editor.MoveSurface(-2, true, false) })
	editor.EditRaiseCeilSlope.NoModifier = true
	editor.EditRaiseCeilSlope.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyRightBracket}
	editor.EditRaiseCeilSlope.Menu = fyne.NewMenuItem("Raise Selection Ceiling Slope", func() { editor.MoveSurface(0.05, false, true) })
	editor.EditLowerCeilSlope.NoModifier = true
	editor.EditLowerCeilSlope.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyLeftBracket}
	editor.EditLowerCeilSlope.Menu = fyne.NewMenuItem("Lower Selection Ceiling Slope", func() { editor.MoveSurface(-0.05, false, true) })
	editor.EditRaiseFloorSlope.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyRightBracket, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditRaiseFloorSlope.Menu = fyne.NewMenuItem("Raise Selection Floor Slope", func() { editor.MoveSurface(0.05, true, true) })
	editor.EditLowerFloorSlope.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyLeftBracket, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditLowerFloorSlope.Menu = fyne.NewMenuItem("Lower Selection Floor Slope", func() { editor.MoveSurface(-0.05, true, true) })
	editor.EditRotateAnchor.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyR, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditRotateAnchor.Menu = fyne.NewMenuItem("Rotate Slope Anchor", func() {
		action := &actions.RotateSegments{IEditor: editor}
		editor.NewAction(action)
		action.Act()
	})

	editor.EditIncreaseGrid.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyUp, Modifier: fyne.KeyModifierShift}
	editor.EditIncreaseGrid.Menu = fyne.NewMenuItem("Increase Grid Size", func() { editor.Current.Step *= 2 })
	editor.EditDecreaseGrid.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyDown, Modifier: fyne.KeyModifierShift}
	editor.EditDecreaseGrid.Menu = fyne.NewMenuItem("Decrease Grid Size", func() { editor.Current.Step /= 2 })

	editor.EditTweakSurfaceLeft.NoModifier = true
	editor.EditTweakSurfaceLeft.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyJ}
	editor.EditTweakSurfaceLeft.Menu = fyne.NewMenuItem("Tweak Surface (Left)", func() {
		t := concepts.IdentityMatrix2
		t.TranslateSelf(&concepts.Vector2{-0.005, 0})
		editor.ChangeSelectedTransformables(&t)
	})

	editor.EditTweakSurfaceRight.NoModifier = true
	editor.EditTweakSurfaceRight.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyL}
	editor.EditTweakSurfaceRight.Menu = fyne.NewMenuItem("Tweak Surface (Right)", func() {
		t := concepts.IdentityMatrix2
		t.TranslateSelf(&concepts.Vector2{0.005, 0})
		editor.ChangeSelectedTransformables(&t)
	})

	editor.EditTweakSurfaceUp.NoModifier = true
	editor.EditTweakSurfaceUp.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyI}
	editor.EditTweakSurfaceUp.Menu = fyne.NewMenuItem("Tweak Surface (Up)", func() {
		t := concepts.IdentityMatrix2
		t.TranslateSelf(&concepts.Vector2{0, 0.005})
		editor.ChangeSelectedTransformables(&t)
	})

	editor.EditTweakSurfaceDown.NoModifier = true
	editor.EditTweakSurfaceDown.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyK}
	editor.EditTweakSurfaceDown.Menu = fyne.NewMenuItem("Tweak Surface (Down)", func() {
		t := concepts.IdentityMatrix2
		t.TranslateSelf(&concepts.Vector2{0, -0.005})
		editor.ChangeSelectedTransformables(&t)
	})

	editor.EditTweakSurfaceCW.NoModifier = true
	editor.EditTweakSurfaceCW.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyO}
	editor.EditTweakSurfaceCW.Menu = fyne.NewMenuItem("Tweak Surface (Clockwise)", func() {
		t := concepts.IdentityMatrix2
		t.RotateSelf(-1)
		editor.ChangeSelectedTransformables(&t)
	})

	editor.EditTweakSurfaceCCW.NoModifier = true
	editor.EditTweakSurfaceCCW.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyU}
	editor.EditTweakSurfaceCCW.Menu = fyne.NewMenuItem("Tweak Surface (Counter-clockwise)", func() {
		t := concepts.IdentityMatrix2
		t.RotateSelf(1)
		editor.ChangeSelectedTransformables(&t)
	})

	editor.ToolsSelect.NoModifier = true
	editor.ToolsSelect.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyEscape}
	editor.ToolsSelect.Menu = fyne.NewMenuItem("Select/Move", func() { editor.SwitchTool(state.ToolSelect) })
	editor.ToolsAddBody.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyB, Modifier: fyne.KeyModifierAlt}
	editor.ToolsAddBody.Menu = fyne.NewMenuItem("Add Body", func() { editor.SwitchTool(state.ToolAddBody) })
	editor.ToolsAddSector.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierAlt}
	editor.ToolsAddSector.Menu = fyne.NewMenuItem("Add Sector", func() { editor.SwitchTool(state.ToolAddSector) })
	editor.ToolsAddInternalSegment.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierAlt | fyne.KeyModifierShift}
	editor.ToolsAddInternalSegment.Menu = fyne.NewMenuItem("Add Internal Segment", func() { editor.SwitchTool(state.ToolAddInternalSegment) })
	editor.ToolsSplitSegment.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyZ, Modifier: fyne.KeyModifierAlt}
	editor.ToolsSplitSegment.Menu = fyne.NewMenuItem("Split Segment", func() { editor.SwitchTool(state.ToolSplitSegment) })
	editor.ToolsSplitSector.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyX, Modifier: fyne.KeyModifierAlt}
	editor.ToolsSplitSector.Menu = fyne.NewMenuItem("Split Sector", func() { editor.SwitchTool(state.ToolSplitSector) })
	editor.ToolsAlignGrid.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyG, Modifier: fyne.KeyModifierAlt}
	editor.ToolsAlignGrid.Menu = fyne.NewMenuItem("Align Grid", func() { editor.SwitchTool(state.ToolAlignGrid) })

	editor.ToolsNewShader.Menu = fyne.NewMenuItem("New Shader...", editor.NewShader)

	editor.ViewSectorEntities.Menu = fyne.NewMenuItem("Toggle Sector Labels", func() { editor.SectorTypesVisible = !editor.SectorTypesVisible })

	editor.BehaviorsReset.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF5, Modifier: fyne.KeyModifierShortcutDefault}
	editor.BehaviorsReset.Menu = fyne.NewMenuItem("Reset all entities", func() {
		editor.DB.Simulation.All.Range(func(key any, _ any) bool {
			d := key.(ecs.Dynamic)
			d.ResetToOriginal()
			if a := d.GetAnimation(); a != nil {
				a.Reset()
			}
			return true
		})
	})
	editor.BehaviorsPause.NoModifier = true
	editor.BehaviorsPause.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF5}
	editor.BehaviorsPause.Menu = fyne.NewMenuItem("Pause simulation", func() {
		editor.DB.EditorPaused = !editor.DB.EditorPaused
		if editor.DB.EditorPaused {
			editor.BehaviorsPause.Menu.Label = "Resume simulation"
		} else {
			editor.BehaviorsPause.Menu.Label = "Pause simulation"
		}
		editor.Window.MainMenu().Items[3].Refresh()
	})

	menuFile := fyne.NewMenu("File", editor.FileOpen.Menu, editor.FileSave.Menu, editor.FileSaveAs.Menu, editor.FileQuit.Menu)
	menuEdit := fyne.NewMenu("Edit", editor.EditUndo.Menu, editor.EditRedo.Menu, fyne.NewMenuItemSeparator(),
		editor.EditCut.Menu, editor.EditCopy.Menu, editor.EditPaste.Menu, editor.EditDelete.Menu, fyne.NewMenuItemSeparator(),
		editor.EditSelectSegment.Menu,
		editor.EditRaiseCeil.Menu, editor.EditLowerCeil.Menu, editor.EditRaiseFloor.Menu, editor.EditLowerFloor.Menu,
		editor.EditRaiseCeilSlope.Menu, editor.EditLowerCeilSlope.Menu, editor.EditRaiseFloorSlope.Menu,
		editor.EditLowerFloorSlope.Menu, editor.EditRotateAnchor.Menu, editor.EditIncreaseGrid.Menu, editor.EditDecreaseGrid.Menu, fyne.NewMenuItemSeparator(),
		editor.EditTweakSurfaceLeft.Menu, editor.EditTweakSurfaceRight.Menu, editor.EditTweakSurfaceUp.Menu, editor.EditTweakSurfaceDown.Menu,
	)
	menuTools := fyne.NewMenu("Tools", editor.ToolsSelect.Menu,
		editor.ToolsAddBody.Menu, editor.ToolsAddSector.Menu, editor.ToolsAddInternalSegment.Menu, editor.ToolsSplitSegment.Menu,
		editor.ToolsSplitSector.Menu, editor.ToolsAlignGrid.Menu, fyne.NewMenuItemSeparator(),
		editor.ToolsNewShader.Menu)

	menuView := fyne.NewMenu("View", editor.ViewSectorEntities.Menu)

	menuBehaviors := fyne.NewMenu("Behaviors", editor.BehaviorsReset.Menu, editor.BehaviorsPause.Menu)

	mainMenu := fyne.NewMainMenu(menuFile, menuEdit, menuTools, menuView, menuBehaviors)
	editor.Window.SetMainMenu(mainMenu)

	editor.MenuActions = make(map[string]*MenuAction)
	allMenus := reflect.ValueOf(&editor.EditorMenu).Elem()
	for i := 0; i < allMenus.NumField(); i++ {
		f := allMenus.Field(i)
		if f.Type() != reflect.TypeOf(MenuAction{}) {
			continue
		}
		action := f.Addr().Interface().(*MenuAction)
		editor.MenuActions[allMenus.Type().Field(i).Name] = action
		if !action.NoModifier && action.Shortcut != nil {
			action.Menu.Shortcut = action.Shortcut
		}
	}
}
