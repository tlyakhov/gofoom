// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"image/color"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"time"

	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/resources"
	_ "tlyakhov/gofoom/scripting_symbols"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/fogleman/gg"
)

var (
	PatternSelectionPrimary   = gg.NewSolidPattern(color.NRGBA{0, 255, 0, 255})
	PatternSelectionSecondary = gg.NewSolidPattern(color.NRGBA{0, 255, 255, 255})
	PatternPVS                = gg.NewSolidPattern(color.NRGBA{160, 255, 160, 255})
	editor                    *Editor
)

var cpuProfile = flag.String("cpuprofile", "", "Write CPU profile to file")

func main() {
	go func() {
		http.ListenAndServe("localhost:8080", nil)
	}()

	flag.Parse()

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}

	ecs.Initialize()
	editor = NewEditor()
	editor.App = app.NewWithID("com.foom.editor")
	editor.App.Lifecycle().SetOnStopped(func() { pprof.StopCPUProfile() })
	editor.Window = editor.App.NewWindow("Foom Editor")
	editor.Window.Resize(fyne.NewSize(1920, 1000))
	editor.Window.CenterOnScreen()

	CreateMainMenu()

	editor.PropertyGrid = container.New(layout.NewFormLayout())
	editor.GridWidget = editor.PropertyGrid
	editor.GridWindow = editor.Window

	editor.EntityList.IEditor = editor
	editor.EntityList.Build()

	scrollProperties := container.NewVScroll(editor.PropertyGrid)
	editor.GameWidget = NewGameWidget()
	editor.MapWidget = NewMapWidget()

	// Create the splitters

	splitMapGame := container.NewVSplit(editor.GameWidget, editor.MapWidget)
	splitMapGame.SetOffset(0.34)
	splitMapGame.Refresh()

	splitEntitiesMap := container.NewHSplit(editor.EntityList.Container, splitMapGame)
	splitEntitiesMap.SetOffset(0.3515)

	splitMapProperties := container.NewHSplit(splitEntitiesMap, scrollProperties)
	splitMapProperties.SetOffset(0.7)

	editor.LabelStatus = widget.NewLabel("")
	editor.LabelStatus.Truncation = fyne.TextTruncateEllipsis

	var item widget.ToolbarItem
	toolbarItems := make([]widget.ToolbarItem, 0)
	item = widget.NewToolbarAction(theme.HomeIcon(), editor.ToolsSelect.Menu.Action)
	toolbarItems = append(toolbarItems, item)
	item = widget.NewToolbarAction(resources.ResourceIconAddEntity, editor.ToolsAddBody.Menu.Action)
	toolbarItems = append(toolbarItems, item)
	item = widget.NewToolbarAction(resources.ResourceIconAddSector, editor.ToolsAddSector.Menu.Action)
	toolbarItems = append(toolbarItems, item)
	item = widget.NewToolbarAction(theme.ContentAddIcon(), editor.ToolsAddInternalSegment.Menu.Action)
	toolbarItems = append(toolbarItems, item)
	item = widget.NewToolbarAction(resources.ResourceIconSplitSector, editor.ToolsSplitSector.Menu.Action)
	toolbarItems = append(toolbarItems, item)
	item = widget.NewToolbarAction(resources.ResourceIconSplitSegment, editor.ToolsSplitSegment.Menu.Action)
	toolbarItems = append(toolbarItems, item)
	item = widget.NewToolbarAction(theme.GridIcon(), editor.ToolsAlignGrid.Menu.Action)
	toolbarItems = append(toolbarItems, item)

	toolbarItems = append(toolbarItems, widget.NewToolbarSeparator())
	item = widget.NewToolbarAction(theme.ContentCutIcon(), editor.EditCut.Menu.Action)
	toolbarItems = append(toolbarItems, item)
	item = widget.NewToolbarAction(theme.ContentCopyIcon(), editor.EditCopy.Menu.Action)
	toolbarItems = append(toolbarItems, item)
	item = widget.NewToolbarAction(theme.ContentPasteIcon(), editor.EditPaste.Menu.Action)
	toolbarItems = append(toolbarItems, item)
	toolbar := widget.NewToolbar(toolbarItems...)

	mainBorder := container.NewBorder(toolbar, editor.LabelStatus, nil, nil, splitMapProperties)
	editor.Window.SetContent(mainBorder)

	editor.App.Lifecycle().SetOnStarted(func() {
		editor.OnStarted()
		editor.Load("data/worlds/hall.yaml")
	})
	editor.App.Lifecycle().SetOnStopped(func() {})

	go func() {
		t := time.NewTicker(time.Second / 60)
		for range t.C {
			ecs.Simulation.Step()
			fyne.DoAndWait(editor.MapWidget.Raster.Refresh)
		}
	}()

	editor.Window.ShowAndRun()
}
