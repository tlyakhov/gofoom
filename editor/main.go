package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"time"

	_ "tlyakhov/gofoom/scripting_symbols"

	"tlyakhov/gofoom/concepts"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
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

func EditorTimer() bool {
	if editor.DB == nil {
		return true
	}
	editor.DB.Simulation.Step()

	return true
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

	editor.App.Lifecycle().SetOnStarted(func() {
		editor.Load("data/worlds/hall.json")
	})

	editor.PropertyGrid = container.New(layout.NewFormLayout())
	editor.FContainer = editor.PropertyGrid

	entities := []any{}
	dataEntities := binding.BindUntypedList(&entities)
	listEntities := widget.NewListWithData(dataEntities, func() fyne.CanvasObject {
		return widget.NewLabel("test")
	}, func(di binding.DataItem, co fyne.CanvasObject) {})

	scrollProperties := container.NewScroll(editor.PropertyGrid)
	editor.GameWidget = NewGameWidget()
	editor.MapWidget = NewMapWidget()
	splitGameProperties := container.NewVSplit(editor.GameWidget, scrollProperties)
	scrollEntities := container.NewScroll(listEntities)
	splitMapGame := container.NewHSplit(editor.MapWidget, splitGameProperties)
	splitMapGame.Refresh()
	splitEntitiesMap := container.NewHSplit(scrollEntities, splitMapGame)
	splitEntitiesMap.SetOffset(0)
	editor.LabelStatus = widget.NewLabel("")
	mainBorder := container.NewBorder(nil, editor.LabelStatus, nil, nil, splitEntitiesMap)
	editor.Window.SetContent(mainBorder)

	go func() {
		for range time.Tick(time.Millisecond * 15) {
			if editor.DB == nil {
				return
			}
			editor.DB.Simulation.Step()
			if editor.MapWidget.NeedsRefresh {
				editor.MapWidget.Raster.Refresh()
				editor.MapWidget.NeedsRefresh = false
			}
		}
	}()

	editor.Window.ShowAndRun()
}
