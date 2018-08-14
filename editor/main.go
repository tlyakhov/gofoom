package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime/pprof"

	"github.com/gotk3/gotk3/glib"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	_ "github.com/tlyakhov/gofoom/behaviors"
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/entities"
	"github.com/tlyakhov/gofoom/logic"
	"github.com/tlyakhov/gofoom/logic/entity"
	"github.com/tlyakhov/gofoom/render"

	_ "github.com/tlyakhov/gofoom/logic/provide"
	_ "github.com/tlyakhov/gofoom/logic/sector"
)

const (
	KEY_LEFT  uint = 65361
	KEY_UP    uint = 65362
	KEY_RIGHT uint = 65363
	KEY_DOWN  uint = 65364

	GridSize                float64 = 10
	SegmentSelectionEpsilon float64 = 5.0
)

var (
	ColorSelectionPrimary   concepts.Vector3 = concepts.Vector3{0, 1, 0}
	ColorSelectionSecondary concepts.Vector3 = concepts.Vector3{0, 1, 1}
	ColorPVS                concepts.Vector3 = concepts.Vector3{0.6, 1, 0.6}
	editor                  *Editor          = NewEditor()
)

func DrawHandle(cr *cairo.Context, v concepts.Vector2) {
	v = editor.WorldToScreen(v)
	v1 := editor.ScreenToWorld(v.Sub(concepts.Vector2{3, 3}))
	v2 := editor.ScreenToWorld(v.Add(concepts.Vector2{3, 3}))
	cr.Rectangle(v1.X, v1.Y, v2.X-v1.X, v2.Y-v1.Y)
	cr.Stroke()
}

func MapTimer(win *gtk.Window) bool {
	editor.GameMap.Frame(15)
	//fmt.Println("timer")
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
	gameArea, _ := gtk.DrawingAreaNew()
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

	renderer := render.NewRenderer()
	renderer.ScreenWidth = 800
	renderer.ScreenHeight = 600
	renderer.Initialize()
	editor.GameMap = logic.LoadMap("data/classicMap.json")
	ps := entity.NewPlayerService(editor.GameMap.Player.(*entities.Player))
	ps.Collide()
	renderer.Map = editor.GameMap.Map
	_, _ = render.NewFont("/Library/Fonts/Courier New.ttf", 24)

	// Event handlers
	mapArea.Connect("draw", DrawMap)

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
			//editor.NewAction(SelectAction{})
		}
		if press.Button() == 2 && editor.CurrentAction == nil {
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
	mapArea.Connect("key-press-event", func(win *gtk.Window, ev *gdk.Event) {
		key := gdk.EventKeyNewFromEvent(ev)
		if key.KeyVal() == gdk.KEY_Shift_L {
			fmt.Println("hmmm")
		}
	})

	glib.TimeoutAdd(15, MapTimer, win)
	gtk.Main()
}
