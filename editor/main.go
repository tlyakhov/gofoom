package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"time"

	"tlyakhov/gofoom/editor/resources"
	_ "tlyakhov/gofoom/scripting_symbols"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	ColorSelectionPrimary   = concepts.Vector3{0, 1, 0}
	ColorSelectionSecondary = concepts.Vector3{0, 1, 1}
	ColorPVS                = concepts.Vector3{0.6, 1, 0.6}
	editor                  *Editor
)

func init() {
	editor = NewEditor()
}

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

	editor.App = app.NewWithID("com.foom.editor")
	editor.Window = editor.App.NewWindow("Foom Editor")
	editor.Window.Resize(fyne.NewSize(1920, 1000))
	editor.Window.CenterOnScreen()

	CreateMainMenu()

	editor.App.Lifecycle().SetOnStarted(func() {
		editor.Load("data/worlds/hall.json")
	})

	editor.PropertyGrid = container.New(layout.NewFormLayout())
	editor.GridWidget = editor.PropertyGrid
	editor.GridWindow = editor.Window

	editor.EntityList.IEditor = editor
	editor.EntityList.Build()

	scrollProperties := container.NewScroll(editor.PropertyGrid)
	editor.GameWidget = NewGameWidget()
	editor.MapWidget = NewMapWidget()
	splitGameProperties := container.NewVSplit(editor.GameWidget, scrollProperties)

	splitMapGame := container.NewHSplit(editor.MapWidget, splitGameProperties)
	splitMapGame.Refresh()
	splitEntitiesMap := container.NewHSplit(editor.EntityList.Container, splitMapGame)
	splitEntitiesMap.SetOffset(0.25)
	editor.LabelStatus = widget.NewLabel("")
	editor.LabelStatus.Truncation = fyne.TextTruncateEllipsis

	var item widget.ToolbarItem
	toolbarItems := make([]widget.ToolbarItem, 0)
	item = widget.NewToolbarAction(theme.HomeIcon(), editor.ToolsSelect.Menu.Action)
	toolbarItems = append(toolbarItems, item)
	item = widget.NewToolbarAction(resources.ResourceIconEntityAdd, editor.ToolsAddBody.Menu.Action)
	toolbarItems = append(toolbarItems, item)
	item = widget.NewToolbarAction(resources.ResourceIconSectorAdd, editor.ToolsAddSector.Menu.Action)
	toolbarItems = append(toolbarItems, item)
	item = widget.NewToolbarAction(theme.GridIcon(), editor.ToolsAlignGrid.Menu.Action)
	toolbarItems = append(toolbarItems, item)

	toolbarItems = append(toolbarItems, widget.NewToolbarSeparator())
	item = widget.NewToolbarAction(theme.ContentCutIcon(), func() {})
	toolbarItems = append(toolbarItems, item)
	item = widget.NewToolbarAction(theme.ContentCopyIcon(), func() {})
	toolbarItems = append(toolbarItems, item)
	item = widget.NewToolbarAction(theme.ContentPasteIcon(), func() {})
	toolbarItems = append(toolbarItems, item)
	toolbar := widget.NewToolbar(toolbarItems...)

	mainBorder := container.NewBorder(toolbar, editor.LabelStatus, nil, nil, splitEntitiesMap)
	editor.Window.SetContent(mainBorder)

	go func() {
		for range time.Tick(time.Millisecond * 15) {
			if editor.DB == nil {
				return
			}
			if editor.Lock.TryRLock() {
				if !editor.SimulationPaused {
					editor.DB.Simulation.Step()
				}
				if editor.MapWidget.NeedsRefresh {
					editor.MapWidget.Raster.Refresh()
					//editor.MapWidget.NeedsRefresh = false
				}
				editor.Lock.RUnlock()
			}
		}
	}()

	editor.Window.ShowAndRun()
}
