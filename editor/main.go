package main

import (
	"flag"
	"log"
	"math"
	"os"
	"runtime/pprof"
	"time"

	"github.com/gotk3/gotk3/glib"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	_ "github.com/tlyakhov/gofoom/behaviors"
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/entities"
	"github.com/tlyakhov/gofoom/logic"
	"github.com/tlyakhov/gofoom/logic/entity"
	"github.com/tlyakhov/gofoom/sectors"

	_ "github.com/tlyakhov/gofoom/logic/provide"
	_ "github.com/tlyakhov/gofoom/logic/sector"
)

const (
	GridSize                float64 = 10
	SegmentSelectionEpsilon float64 = 5.0
)

var (
	ColorSelectionPrimary   = concepts.Vector3{0, 1, 0}
	ColorSelectionSecondary = concepts.Vector3{0, 1, 1}
	ColorPVS                = concepts.Vector3{0.6, 1, 0.6}
	editor                  = NewEditor()
	gameKeyMap              = make(map[uint]bool)
	last                    = time.Now()
)

func EditorTimer(win *gtk.Window) bool {
	dt := time.Since(last).Seconds() * 1000
	last = time.Now()

	editor.GameMap.Frame(dt)
	editor.GatherHoveringObjects()

	ps := entity.NewPlayerService(editor.GameMap.Player.(*entities.Player))

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

var cpuProfile = flag.String("cpuprofile", "", "Write CPU profile to file")

func main() {
	flag.Parse()

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	gtk.Init(nil)

	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	mapArea, _ := gtk.DrawingAreaNew()
	mapArea.SetEvents(int(gdk.POINTER_MOTION_MASK) | int(gdk.SCROLL_MASK) | int(gdk.BUTTON_PRESS_MASK) | int(gdk.BUTTON_RELEASE_MASK) | int(gdk.KEY_PRESS_MASK))
	mapArea.SetCanFocus(true)
	gameArea, _ := gtk.DrawingAreaNew()
	gameArea.SetEvents(int(gdk.KEY_PRESS_MASK) | int(gdk.KEY_RELEASE_MASK))
	gameArea.SetCanFocus(true)
	propGrid, _ := gtk.GridNew()
	hpane, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	vpane, _ := gtk.PanedNew(gtk.ORIENTATION_VERTICAL)
	hpane.Pack1(mapArea, true, true)
	hpane.Pack2(vpane, true, true)
	vpane.Pack1(gameArea, true, true)
	vpane.Pack2(propGrid, true, true)
	win.Add(hpane)
	win.SetSizeRequest(1280, 720)
	win.SetPosition(gtk.WIN_POS_CENTER)
	win.SetTitle("Foom Editor")
	win.Connect("destroy", gtk.MainQuit)
	win.ShowAll()

	editor.GameMap = logic.LoadMap("data/classicMap.json")
	ps := entity.NewPlayerService(editor.GameMap.Player.(*entities.Player))
	ps.Collide()

	editor.GameView(gameArea.GetAllocatedWidth(), gameArea.GetAllocatedHeight())

	// Event handlers
	mapArea.Connect("draw", DrawMap)
	gameArea.Connect("draw", DrawGame)

	mapArea.Connect("motion-notify-event", func(da *gtk.DrawingArea, ev *gdk.Event) {
		motion := gdk.EventMotionNewFromEvent(ev)
		x, y := motion.MotionVal()
		if x == editor.Mouse.X && y == editor.Mouse.Y {
			return
		}
		editor.Mouse.X = x
		editor.Mouse.Y = y
		editor.MouseWorld = editor.ScreenToWorld(editor.Mouse)

		if editor.CurrentAction != nil {
			editor.CurrentAction.OnMouseMove()
		}
	})

	mapArea.Connect("button-press-event", func(da *gtk.DrawingArea, ev *gdk.Event) {
		press := gdk.EventButtonNewFromEvent(ev)
		editor.MousePressed = true
		editor.MouseDown.X, editor.MouseDown.Y = press.MotionVal()
		editor.MouseDownWorld = editor.ScreenToWorld(editor.MouseDown)

		if press.Button() == 3 && editor.CurrentAction == nil {
			editor.NewAction(&SelectAction{Editor: editor})
		} else if press.Button() == 2 && editor.CurrentAction == nil {
			editor.NewAction(&PanAction{Editor: editor})
		}

		if editor.CurrentAction != nil {
			editor.CurrentAction.OnMouseDown(press)
		}
	})

	mapArea.Connect("button-release-event", func(da *gtk.DrawingArea, ev *gdk.Event) {
		//release := &gdk.EventButton{ev}
		editor.MousePressed = false

		if editor.CurrentAction != nil {
			editor.CurrentAction.OnMouseUp()
		}
	})

	mapArea.Connect("scroll-event", func(da *gtk.DrawingArea, ev *gdk.Event) {
		scroll := gdk.EventScrollNewFromEvent(ev)
		delta := math.Abs(scroll.DeltaY() / 5)
		if scroll.Direction() == gdk.SCROLL_DOWN {
			delta = -delta
		}
		if editor.Scale > 0.25 {
			editor.Scale += delta * 0.2
		} else if editor.Scale > 0.025 {
			editor.Scale += delta * 0.02
		} else if editor.Scale > 0.0025 {
			editor.Scale += delta * 0.002
		}
	})
	mapArea.Connect("key-press-event", func(da *gtk.DrawingArea, ev *gdk.Event) {
		//key := gdk.EventKeyNewFromEvent(ev)
	})
	gameArea.Connect("key-press-event", func(da *gtk.DrawingArea, ev *gdk.Event) {
		key := gdk.EventKeyNewFromEvent(ev)
		gameKeyMap[key.KeyVal()] = true
	})
	gameArea.Connect("key-release-event", func(da *gtk.DrawingArea, ev *gdk.Event) {
		key := gdk.EventKeyNewFromEvent(ev)
		delete(gameKeyMap, key.KeyVal())
	})

	glib.TimeoutAdd(15, EditorTimer, win)
	gtk.Main()
}
