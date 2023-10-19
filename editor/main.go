package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"time"

	"tlyakhov/gofoom/editor/actions"

	"github.com/gotk3/gotk3/glib"

	_ "tlyakhov/gofoom/behaviors"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/editor/state"
	"tlyakhov/gofoom/entities"
	"tlyakhov/gofoom/logic/entity"
	"tlyakhov/gofoom/sectors"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	_ "tlyakhov/gofoom/logic/provide"
	_ "tlyakhov/gofoom/logic/sector"
)

var (
	ColorSelectionPrimary   = concepts.Vector3{0, 1, 0}
	ColorSelectionSecondary = concepts.Vector3{0, 1, 1}
	ColorPVS                = concepts.Vector3{0.6, 1, 0.6}
	editor                  = NewEditor()
	gameKeyMap              = make(map[uint]bool)
	last                    = time.Now()
)

func EditorTimer() bool {
	dt := time.Since(last).Seconds() * 1000
	editor.LastFrameTime = dt
	last = time.Now()

	ps := entity.NewPlayerService(editor.World.Player.(*entities.Player))

	if gameKeyMap[gdk.KEY_w] {
		ps.Move(ps.Player.Angle, 1.0)
	}
	if gameKeyMap[gdk.KEY_s] {
		ps.Move(ps.Player.Angle+180.0, 1.0)
	}
	if gameKeyMap[gdk.KEY_e] {
		ps.Move(ps.Player.Angle+90.0, 0.5)
	}
	if gameKeyMap[gdk.KEY_q] {
		ps.Move(ps.Player.Angle+270.0, 0.5)
	}
	if gameKeyMap[gdk.KEY_a] {
		ps.Player.Angle -= constants.PlayerTurnSpeed / 30.0
		ps.Player.Angle = concepts.NormalizeAngle(ps.Player.Angle)
	}
	if gameKeyMap[gdk.KEY_d] {
		ps.Player.Angle += constants.PlayerTurnSpeed / 30.0
		ps.Player.Angle = concepts.NormalizeAngle(ps.Player.Angle)
	}
	if gameKeyMap[gdk.KEY_space] {
		if _, ok := ps.Player.Sector.(*sectors.Underwater); ok {
			ps.Player.Vel.Now[2] += constants.PlayerSwimStrength * constants.TimeStep
		} else if ps.Player.OnGround {
			ps.Player.Vel.Now[2] += constants.PlayerJumpStrength * constants.TimeStep
			ps.Player.OnGround = false
		}
	}
	if gameKeyMap[gdk.KEY_c] {
		if _, ok := ps.Player.Sector.(*sectors.Underwater); ok {
			ps.Player.Vel.Now[2] -= constants.PlayerSwimStrength * constants.TimeStep
		} else {
			ps.Crouching = true
		}
	} else {
		ps.Crouching = false
	}

	editor.Window.QueueDraw()
	return true
}

func setupMenu() {
	menuBuilder, err := gtk.BuilderNew()
	if err != nil {
		log.Fatal("Can't create GTK+ builder.", err)
	}
	err = menuBuilder.AddFromFile("editor-menu.ui")
	if err != nil {
		log.Fatal("Can't load GTK+ Menu UI from file editor-menu.ui.", err)
	}
	menu, err := menuBuilder.GetObject("Menu")
	if err != nil {
		log.Fatal("Can't find Menu object in GTK+ menu UI file.", err)
	}

	editor.App.SetMenubar(menu.(*glib.MenuModel))
}

func onActivate() {
	builder, err := gtk.BuilderNew()
	if err != nil {
		log.Fatal("Can't create GTK+ builder.", err)
	}
	err = builder.AddFromFile("editor-glade.glade")
	if err != nil {
		log.Fatal("Can't load GTK+ UI from file.", err)
	}
	obj, err := builder.GetObject("MainWindow")
	if err != nil {
		log.Fatal("Can't find MainWindow object in GTK+ UI file.", err)
	}
	editor.Window = obj.(*gtk.ApplicationWindow)
	obj, err = builder.GetObject("MapArea")
	if err != nil {
		log.Fatal("Can't find MapArea object in GTK+ UI file.", err)
	}
	editor.MapArea = obj.(*gtk.DrawingArea)
	obj, err = builder.GetObject("GameArea")
	if err != nil {
		log.Fatal("Can't find GameArea object in GTK+ UI file.", err)
	}
	editor.GameArea = obj.(*gtk.DrawingArea)
	obj, err = builder.GetObject("PropertyGrid")
	if err != nil {
		log.Fatal("Can't find PropertyGrid object in GTK+ UI file.", err)
	}
	editor.Grid.Container = obj.(*gtk.Grid)
	obj, err = builder.GetObject("EntityTypes")
	if err != nil {
		log.Fatal("Can't find EntityTypes object in GTK+ UI file.", err)
	}
	editor.EntityTypes = obj.(*gtk.ComboBoxText)
	obj, err = builder.GetObject("SectorTypes")
	if err != nil {
		log.Fatal("Can't find SectorTypes object in GTK+ UI file.", err)
	}
	editor.SectorTypes = obj.(*gtk.ComboBoxText)
	obj, err = builder.GetObject("StatusBar")
	if err != nil {
		log.Fatal("Can't find StatusBar object in GTK+ UI file.", err)
	}
	editor.StatusBar = obj.(*gtk.Label)

	editor.AddSimpleMenuAction("open", MainOpen)
	editor.AddSimpleMenuAction("save", MainSave)
	editor.AddSimpleMenuAction("saveas", MainSaveAs)
	editor.AddSimpleMenuAction("quit", func(obj *glib.Object) { editor.App.Quit() })
	editor.AddSimpleMenuAction("undo", func(obj *glib.Object) { editor.Undo() })
	editor.AddSimpleMenuAction("redo", func(obj *glib.Object) { editor.Redo() })
	editor.AddSimpleMenuAction("delete", func(obj *glib.Object) {
		action := &actions.Delete{IEditor: editor}
		editor.NewAction(action)
		action.Act()
	})
	editor.AddSimpleMenuAction("tool.select", func(obj *glib.Object) {
		editor.SwitchTool(state.ToolSelect)
		tool, _ := builder.GetObject("ToolSelect")
		tool.(*gtk.RadioButton).SetProperty("active", true)
	})
	editor.AddSimpleMenuAction("tool.raise.ceil", func(obj *glib.Object) { editor.MoveSurface(2, false, false) })
	editor.AddSimpleMenuAction("tool.lower.ceil", func(obj *glib.Object) { editor.MoveSurface(-2, false, false) })
	editor.AddSimpleMenuAction("tool.raise.floor", func(obj *glib.Object) { editor.MoveSurface(2, true, false) })
	editor.AddSimpleMenuAction("tool.lower.floor", func(obj *glib.Object) { editor.MoveSurface(-2, true, false) })
	editor.AddSimpleMenuAction("tool.raise.ceil.slope", func(obj *glib.Object) { editor.MoveSurface(0.05, false, true) })
	editor.AddSimpleMenuAction("tool.lower.ceil.slope", func(obj *glib.Object) { editor.MoveSurface(-0.05, false, true) })
	editor.AddSimpleMenuAction("tool.raise.floor.slope", func(obj *glib.Object) { editor.MoveSurface(0.05, true, true) })
	editor.AddSimpleMenuAction("tool.lower.floor.slope", func(obj *glib.Object) { editor.MoveSurface(-0.05, true, true) })
	editor.AddSimpleMenuAction("tool.rotate.slope", func(obj *glib.Object) {
		action := &actions.RotateSegments{IEditor: editor}
		editor.NewAction(action)
		action.Act()
	})
	editor.AddSimpleMenuAction("tool.grid.up", func(obj *glib.Object) {
		editor.Current.Step *= 2
	})
	editor.AddSimpleMenuAction("tool.grid.down", func(obj *glib.Object) { editor.Current.Step /= 2 })

	setupMenu()
	editor.Window.SetApplication(editor.App)
	editor.Window.ShowAll()
	editor.Window.Maximize()

	// Event handlers
	signals := make(map[string]interface{})
	signals["Menu.File.Quit"] = func(obj *glib.Object) {
		editor.App.Quit()
	}
	signals["MapArea.Draw"] = DrawMap
	signals["MapArea.MotionNotify"] = MapMotionNotify
	signals["MapArea.ButtonPress"] = MapButtonPress
	signals["MapArea.ButtonRelease"] = MapButtonRelease
	signals["GameArea.ButtonRelease"] = GameButtonPress
	signals["GameArea.Draw"] = DrawGame
	signals["GameArea.ButtonPress"] = GameButtonPress
	signals["GameArea.KeyPress"] = func(da *gtk.DrawingArea, ev *gdk.Event) {
		key := gdk.EventKeyNewFromEvent(ev)
		gameKeyMap[key.KeyVal()] = true
	}
	signals["GameArea.KeyRelease"] = func(da *gtk.DrawingArea, ev *gdk.Event) {
		key := gdk.EventKeyNewFromEvent(ev)
		delete(gameKeyMap, key.KeyVal())
	}
	signals["MapArea.Scroll"] = MapScroll
	signals["Tools.Toggled"] = func(obj *glib.Object) {
		active, _ := obj.GetProperty("active")
		if !active.(bool) {
			return
		}

		// We don't have gtk.Buildable available so we can't get the ids. :(
		switch label, _ := obj.GetProperty("label"); label {
		case "Select/Move":
			editor.SwitchTool(state.ToolSelect)
		case "Split Segment":
			editor.SwitchTool(state.ToolSplitSegment)
		case "Split Sector":
			editor.SwitchTool(state.ToolSplitSector)
		case "Add Sector":
			editor.SwitchTool(state.ToolAddSector)
		case "Add Entity":
			editor.SwitchTool(state.ToolAddEntity)
		case "Align Grid":
			editor.SwitchTool(state.ToolAlignGrid)
		}
	}
	builder.ConnectSignals(signals)

	editor.Load("data/worlds/empty.json")
	//editor.Test()
	glib.TimeoutAdd(15, EditorTimer)
}

var cpuProfile = flag.String("cpuprofile", "", "Write CPU profile to file")

func main() {
	go func() {
		http.ListenAndServe("localhost:8080", nil)
	}()

	gtk.Init(&os.Args)
	// Gtk bindings are missing this widget in builders. Add it in manually.
	gtk.WrapMap["GtkRadioToolButton"] = gtk.WrapMap["GtkRadioButton"]
	flag.Parse()

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	const appID = "com.foom.editor"
	var err error
	editor.App, err = gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)

	if err != nil {
		log.Fatal("Could not create application.", err)
	}

	editor.App.Connect("activate", onActivate)
	os.Exit(editor.App.Run(os.Args))
}
