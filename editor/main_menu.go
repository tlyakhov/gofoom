// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"log"
	"reflect"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type MenuAction struct {
	Shortcut   fyne.Shortcut
	Menu       *fyne.MenuItem
	NoModifier bool
}

type EditorMenu struct {
	FileOpen   MenuAction
	FileSaveAs MenuAction
	FileSave   MenuAction
	FileQuit   MenuAction

	EditUndo        MenuAction
	EditRedo        MenuAction
	EditDelete      MenuAction
	EditCut         MenuAction
	EditCopy        MenuAction
	EditPaste       MenuAction
	EditFindReplace MenuAction

	EditSelectSegment         MenuAction
	EditRaiseCeil             MenuAction
	EditLowerCeil             MenuAction
	EditRaiseFloor            MenuAction
	EditLowerFloor            MenuAction
	EditRotateCeilAzimuthCW   MenuAction
	EditRotateCeilAzimuthCCW  MenuAction
	EditRotateFloorAzimuthCW  MenuAction
	EditRotateFloorAzimuthCCW MenuAction
	EditRotateAnchor          MenuAction
	EditIncreaseGrid          MenuAction
	EditDecreaseGrid          MenuAction

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
	ToolsPathDebug          MenuAction

	ViewSectorEntities     MenuAction
	ViewSnapToGrid         MenuAction
	ViewDisabledProperties MenuAction

	BehaviorsPause   MenuAction
	BehaviorsReset   MenuAction
	BehaviorsRespawn MenuAction

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
			ecs.Save(uc.URI().Path())
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
		ecs.Save(editor.OpenFile)
		editor.Modified = false
		editor.UpdateTitle()
	})
	editor.FileQuit.Menu = fyne.NewMenuItem("Quit", func() {})

	editor.EditUndo.Shortcut = &fyne.ShortcutUndo{}
	editor.EditUndo.Menu = fyne.NewMenuItem("Undo", func() {
		editor.FocusedShortcut(editor.EditUndo.Shortcut)
	})
	editor.EditRedo.Shortcut = &fyne.ShortcutRedo{}
	editor.EditRedo.Menu = fyne.NewMenuItem("Redo", func() {
		editor.FocusedShortcut(editor.EditRedo.Shortcut)
	})

	editor.EditDelete.NoModifier = true
	editor.EditDelete.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyDelete, Modifier: 0}
	editor.EditDelete.Menu = fyne.NewMenuItem("Delete", func() {
		editor.Act(&actions.Delete{Action: state.Action{IEditor: editor}})
	})

	editor.EditCut.Shortcut = &fyne.ShortcutCut{Clipboard: editor.App.Clipboard()}
	editor.EditCut.Menu = fyne.NewMenuItem("Cut", func() {
		editor.FocusedShortcut(editor.EditCut.Shortcut)
	})

	editor.EditCopy.Shortcut = &fyne.ShortcutCopy{Clipboard: editor.App.Clipboard()}
	editor.EditCopy.Menu = fyne.NewMenuItem("Copy", func() {
		editor.FocusedShortcut(editor.EditCopy.Shortcut)
	})

	editor.EditPaste.Shortcut = &fyne.ShortcutPaste{Clipboard: editor.App.Clipboard()}
	editor.EditPaste.Menu = fyne.NewMenuItem("Paste", func() {
		editor.FocusedShortcut(editor.EditPaste.Shortcut)

	})

	editor.EditFindReplace.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF, Modifier: fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift}
	editor.EditFindReplace.Menu = fyne.NewMenuItem("Find & Replace all references", func() {
		fromEntry := widget.NewEntry()
		fromEntry.Text = ""
		fromEntry.PlaceHolder = "e.g. 123"
		toEntry := widget.NewEntry()
		toEntry.Text = ""
		toEntry.PlaceHolder = "e.g. 123"

		title := "Find & Replace all references"

		dialog.ShowForm(title, "Find & Replace", "Cancel", []*widget.FormItem{
			{Text: "From", Widget: fromEntry},
			{Text: "To", Widget: toEntry},
		}, func(b bool) {
			if !b {
				return
			}
			fromEntity, err := ecs.ParseEntityHumanOrCanonical(fromEntry.Text)
			if err != nil {
				log.Printf("Error: %v", err)
				return
			}
			toEntity, err := ecs.ParseEntityHumanOrCanonical(toEntry.Text)
			if err != nil {
				log.Printf("Error: %v", err)
				return
			}
			if fromEntity.Local() <= 1 || toEntity.Local() <= 1 {
				log.Printf("Can't find/replace entity 0 or 1")
				return
			}
			editor.Act(&actions.FindReplace{
				Action: state.Action{IEditor: editor},
				From:   fromEntity, To: toEntity})
		}, editor.GridWindow)

	})

	// TODO: Select all would be nice as well

	editor.EditSelectSegment.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyApostrophe, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditSelectSegment.Menu = fyne.NewMenuItem("Select First/Next Segment", editor.ToolSelectSegment)
	editor.EditRaiseCeil.NoModifier = true
	editor.EditRaiseCeil.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyPageUp}
	editor.EditRaiseCeil.Menu = fyne.NewMenuItem("Raise Selection Ceiling", func() {
		editor.Act(&actions.MoveSurface{Action: state.Action{IEditor: editor}, Mode: actions.SurfaceModeLevel, Delta: 2, Floor: false})
	})
	editor.EditLowerCeil.NoModifier = true
	editor.EditLowerCeil.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyPageDown}
	editor.EditLowerCeil.Menu = fyne.NewMenuItem("Lower Selection Ceiling", func() {
		editor.Act(&actions.MoveSurface{Action: state.Action{IEditor: editor}, Mode: actions.SurfaceModeLevel, Delta: -2, Floor: false})
	})
	editor.EditRaiseFloor.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyPageUp, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditRaiseFloor.Menu = fyne.NewMenuItem("Raise Selection Floor", func() {
		editor.Act(&actions.MoveSurface{Action: state.Action{IEditor: editor}, Mode: actions.SurfaceModeLevel, Delta: 2, Floor: true})
	})
	editor.EditLowerFloor.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyPageDown, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditLowerFloor.Menu = fyne.NewMenuItem("Lower Selection Floor", func() {
		editor.Act(&actions.MoveSurface{Action: state.Action{IEditor: editor}, Mode: actions.SurfaceModeLevel, Delta: -2, Floor: true})
	})
	editor.EditRotateCeilAzimuthCW.NoModifier = true
	editor.EditRotateCeilAzimuthCW.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyRightBracket}
	editor.EditRotateCeilAzimuthCW.Menu = fyne.NewMenuItem("Rotate Selection Ceiling Azimuth CW", func() {
		editor.Act(&actions.MoveSurface{Action: state.Action{IEditor: editor}, Mode: actions.SurfaceModePhi, Delta: 1, Floor: false})
	})
	editor.EditRotateCeilAzimuthCCW.NoModifier = true
	editor.EditRotateCeilAzimuthCCW.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyLeftBracket}
	editor.EditRotateCeilAzimuthCCW.Menu = fyne.NewMenuItem("Rotate Selection Ceiling Azimuth CCW", func() {
		editor.Act(&actions.MoveSurface{Action: state.Action{IEditor: editor}, Mode: actions.SurfaceModePhi, Delta: -1, Floor: false})
	})

	editor.EditRotateFloorAzimuthCW.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyRightBracket, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditRotateFloorAzimuthCW.Menu = fyne.NewMenuItem("Rotate Selection Floor Azimuth CW", func() {
		editor.Act(&actions.MoveSurface{Action: state.Action{IEditor: editor}, Mode: actions.SurfaceModePhi, Delta: 1, Floor: true})
	})
	editor.EditRotateFloorAzimuthCCW.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyLeftBracket, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditRotateFloorAzimuthCCW.Menu = fyne.NewMenuItem("Rotate Selection Floor Azimuth CCW", func() {
		editor.Act(&actions.MoveSurface{Action: state.Action{IEditor: editor}, Mode: actions.SurfaceModePhi, Delta: -1, Floor: true})
	})

	// TODO: Instead, make two actions, to match floor/ceiling slope to a segment
	editor.EditRotateAnchor.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyR, Modifier: fyne.KeyModifierShortcutDefault}
	editor.EditRotateAnchor.Menu = fyne.NewMenuItem("Rotate Slope Anchor", func() {
		editor.Act(&actions.RotateSegments{Action: state.Action{IEditor: editor}})
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
	editor.ToolsPathDebug.Menu = fyne.NewMenuItem("Path Debug", func() { editor.SwitchTool(state.ToolPathDebug) })

	editor.ViewSectorEntities.Menu = fyne.NewMenuItem("Toggle Sector Labels", func() {
		editor.SectorTypesVisible = !editor.SectorTypesVisible
		editor.ViewSectorEntities.Menu.Checked = editor.SectorTypesVisible
	})
	editor.ViewSectorEntities.Menu.Checked = editor.SectorTypesVisible

	editor.ViewSnapToGrid.Menu = fyne.NewMenuItem("Toggle Snap to Grid", func() {
		editor.Snap = !editor.Snap
		editor.ViewSnapToGrid.Menu.Checked = editor.Snap
	})
	editor.ViewSnapToGrid.Menu.Checked = editor.Snap

	editor.ViewDisabledProperties.Menu = fyne.NewMenuItem("Toggle disabled properties", func() {
		editor.DisabledPropertiesVisible = !editor.DisabledPropertiesVisible
		editor.ViewDisabledProperties.Menu.Checked = editor.DisabledPropertiesVisible
	})
	editor.ViewDisabledProperties.Menu.Checked = editor.DisabledPropertiesVisible

	editor.BehaviorsReset.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF5, Modifier: fyne.KeyModifierShortcutDefault}
	editor.BehaviorsReset.Menu = fyne.NewMenuItem("Reset all entities", func() { controllers.ResetAllSpawnables() })
	editor.BehaviorsPause.NoModifier = true
	editor.BehaviorsPause.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF5}
	editor.BehaviorsPause.Menu = fyne.NewMenuItem("Pause simulation", func() {
		ecs.Simulation.EditorPaused = !ecs.Simulation.EditorPaused
		if ecs.Simulation.EditorPaused {
			editor.BehaviorsPause.Menu.Label = "Resume simulation"
		} else {
			editor.BehaviorsPause.Menu.Label = "Pause simulation"
		}
		fyne.Do(editor.Window.MainMenu().Items[3].Refresh)
	})
	editor.BehaviorsRespawn.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF5, Modifier: fyne.KeyModifierAlt}
	editor.BehaviorsRespawn.Menu = fyne.NewMenuItem("Respawn", func() {
		editor.Lock.Lock()
		defer editor.Lock.Unlock()
		controllers.RespawnAll()
	})

	menuFile := fyne.NewMenu("File", editor.FileOpen.Menu, editor.FileSave.Menu, editor.FileSaveAs.Menu, editor.FileQuit.Menu)
	menuEdit := fyne.NewMenu("Edit", editor.EditUndo.Menu, editor.EditRedo.Menu, fyne.NewMenuItemSeparator(),
		editor.EditCut.Menu, editor.EditCopy.Menu, editor.EditPaste.Menu, editor.EditDelete.Menu, editor.EditFindReplace.Menu, fyne.NewMenuItemSeparator(),
		editor.EditSelectSegment.Menu,
		editor.EditRaiseCeil.Menu, editor.EditLowerCeil.Menu, editor.EditRaiseFloor.Menu, editor.EditLowerFloor.Menu,
		editor.EditRotateCeilAzimuthCW.Menu, editor.EditRotateCeilAzimuthCCW.Menu, editor.EditRotateFloorAzimuthCW.Menu,
		editor.EditRotateFloorAzimuthCCW.Menu, editor.EditRotateAnchor.Menu, editor.EditIncreaseGrid.Menu, editor.EditDecreaseGrid.Menu, fyne.NewMenuItemSeparator(),
		editor.EditTweakSurfaceLeft.Menu, editor.EditTweakSurfaceRight.Menu, editor.EditTweakSurfaceUp.Menu, editor.EditTweakSurfaceDown.Menu,
	)
	menuTools := fyne.NewMenu("Tools", editor.ToolsSelect.Menu,
		editor.ToolsAddBody.Menu, editor.ToolsAddSector.Menu, editor.ToolsAddInternalSegment.Menu, editor.ToolsSplitSegment.Menu,
		editor.ToolsSplitSector.Menu, editor.ToolsAlignGrid.Menu, fyne.NewMenuItemSeparator(),
		editor.ToolsNewShader.Menu, editor.ToolsPathDebug.Menu)

	menuView := fyne.NewMenu("View", editor.ViewSectorEntities.Menu, editor.ViewSnapToGrid.Menu)

	menuBehaviors := fyne.NewMenu("Behaviors", editor.BehaviorsReset.Menu, editor.BehaviorsPause.Menu, editor.BehaviorsRespawn.Menu)

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
