package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/tlyakhov/gofoom/editor/actions"

	"github.com/gotk3/gotk3/glib"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	_ "github.com/tlyakhov/gofoom/behaviors"
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/editor/state"
	"github.com/tlyakhov/gofoom/entities"
	"github.com/tlyakhov/gofoom/logic/entity"
	"github.com/tlyakhov/gofoom/sectors"

	_ "github.com/tlyakhov/gofoom/logic/provide"
	_ "github.com/tlyakhov/gofoom/logic/sector"
)

var (
	ColorSelectionPrimary   = concepts.Vector3{0, 1, 0}
	ColorSelectionSecondary = concepts.Vector3{0, 1, 1}
	ColorPVS                = concepts.Vector3{0.9, 1, 0.9}
	editor                  = NewEditor()
	gameKeyMap              = make(map[uint]bool)
	last                    = time.Now()
)

func EditorTimer(win *gtk.ApplicationWindow) bool {
	dt := time.Since(last).Seconds() * 1000
	last = time.Now()

	editor.World.Frame(dt)
	editor.GatherHoveringObjects()

	ps := entity.NewPlayerService(editor.World.Player.(*entities.Player))

	if gameKeyMap[gdk.KEY_w] {
		ps.Move(ps.Player.Angle, dt, 1.0)
	}
	if gameKeyMap[gdk.KEY_s] {
		ps.Move(ps.Player.Angle+180.0, dt, 1.0)
	}
	if gameKeyMap[gdk.KEY_e] {
		ps.Move(ps.Player.Angle+90.0, dt, 0.5)
	}
	if gameKeyMap[gdk.KEY_q] {
		ps.Move(ps.Player.Angle+270.0, dt, 0.5)
	}
	if gameKeyMap[gdk.KEY_a] {
		ps.Player.Angle -= constants.PlayerTurnSpeed * dt / 30.0
		ps.Player.Angle = concepts.NormalizeAngle(ps.Player.Angle)
	}
	if gameKeyMap[gdk.KEY_d] {
		ps.Player.Angle += constants.PlayerTurnSpeed * dt / 30.0
		ps.Player.Angle = concepts.NormalizeAngle(ps.Player.Angle)
	}
	if gameKeyMap[gdk.KEY_space] {
		if _, ok := ps.Player.Sector.(*sectors.Underwater); ok {
			ps.Player.Vel.Z += constants.PlayerSwimStrength * dt / 30.0
		} else if ps.Standing {
			ps.Player.Vel.Z += constants.PlayerJumpStrength * dt / 30.0
			ps.Standing = false
		}
	}
	if gameKeyMap[gdk.KEY_c] {
		if _, ok := ps.Player.Sector.(*sectors.Underwater); ok {
			ps.Player.Vel.Z -= constants.PlayerSwimStrength * dt / 30.0
		} else {
			ps.Crouching = true
		}
	} else {
		ps.Crouching = false
	}

	win.QueueDraw()
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

	editor.App.SetMenubar(&menu.(*glib.Menu).MenuModel)
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

	setupMenu()
	editor.Window.SetApplication(editor.App)
	editor.Window.ShowAll()

	// Event handlers
	signals := make(map[string]interface{})
	signals["Menu.File.Quit"] = func(obj *glib.Object) {
		editor.App.Quit()
	}
	signals["MapArea.Draw"] = DrawMap
	signals["MapArea.MotionNotify"] = MapMotionNotify
	signals["MapArea.ButtonPress"] = MapButtonPress
	signals["MapArea.ButtonRelease"] = MapButtonRelease
	signals["GameArea.Draw"] = DrawGame
	signals["GameArea.ButtonPress"] = func(da *gtk.DrawingArea, ev *gdk.Event) {
		da.GrabFocus()
	}
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
	glib.TimeoutAdd(15, EditorTimer, editor.Window)
}

var cpuProfile = flag.String("cpuprofile", "", "Write CPU profile to file")

func main() {
	gtk.Init(&os.Args)
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
