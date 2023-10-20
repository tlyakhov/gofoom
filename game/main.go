package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"reflect"
	"runtime/pprof"

	_ "tlyakhov/gofoom/behaviors"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/controllers/entity"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/entities"
	"tlyakhov/gofoom/sectors"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	_ "tlyakhov/gofoom/controllers/provide"
	_ "tlyakhov/gofoom/controllers/sector"
	"tlyakhov/gofoom/render"

	// "math"
	// "math/rand"
	// "os"

	"image/color"
	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var cpuProfile = flag.String("cpuprofile", "", "Write CPU profile to file")
var win *pixelgl.Window
var ps *entity.PlayerService
var sim *core.Simulation
var renderer *render.Renderer
var gameMap *controllers.MapController
var canvas *pixelgl.Canvas
var buffer *image.RGBA
var mainFont *render.Font

func processInput() {
	win.SetClosed(win.JustPressed(pixelgl.KeyEscape))

	if win.JustPressed(pixelgl.MouseButtonLeft) {
	}
	if win.Pressed(pixelgl.KeyW) {
		ps.Move(ps.Player.Angle)
	}
	if win.Pressed(pixelgl.KeyS) {
		ps.Move(ps.Player.Angle + 180.0)
	}
	if win.Pressed(pixelgl.KeyE) {
		ps.Move(ps.Player.Angle + 90.0)
	}
	if win.Pressed(pixelgl.KeyQ) {
		ps.Move(ps.Player.Angle + 270.0)
	}
	if win.Pressed(pixelgl.KeyA) {
		ps.Player.Angle -= constants.PlayerTurnSpeed * constants.TimeStep
		ps.Player.Angle = concepts.NormalizeAngle(ps.Player.Angle)
	}
	if win.Pressed(pixelgl.KeyD) {
		ps.Player.Angle += constants.PlayerTurnSpeed * constants.TimeStep
		ps.Player.Angle = concepts.NormalizeAngle(ps.Player.Angle)
	}
	if win.Pressed(pixelgl.KeySpace) {
		if _, ok := ps.Player.Sector.(*sectors.Underwater); ok {
			ps.Player.Vel.Now[2] += constants.PlayerSwimStrength * constants.TimeStep
		} else if ps.Player.OnGround {
			ps.Player.Vel.Now[2] += constants.PlayerJumpStrength * constants.TimeStep
			ps.Player.OnGround = false
		}
	}
	if win.Pressed(pixelgl.KeyC) {
		if _, ok := ps.Player.Sector.(*sectors.Underwater); ok {
			ps.Player.Vel.Now[2] -= constants.PlayerSwimStrength * constants.TimeStep
		} else {
			ps.Crouching = true
		}
	} else {
		ps.Crouching = false
	}
}

func integrateGame() {
	processInput()
	gameMap.Frame()
}

func renderGame() {
	renderer.Render(buffer.Pix)
	canvas.SetPixels(buffer.Pix)
	winw := win.Bounds().W()
	winh := win.Bounds().H()
	mat := pixel.IM.ScaledXY(pixel.Vec{X: 0, Y: 0}, pixel.Vec{X: winw / float64(renderer.ScreenWidth), Y: -winh / float64(renderer.ScreenHeight)}).Moved(win.Bounds().Center())
	canvas.Draw(win, mat)
	mainFont.Draw(win, 10, 10, color.NRGBA{0xff, 0, 0, 0xff}, fmt.Sprintf("FPS: %.1f", sim.FPS))
	mainFont.Draw(win, 10, 20, color.NRGBA{0xff, 0, 0, 0xff}, fmt.Sprintf("Health: %.1f", ps.Player.Health))
	mainFont.Draw(win, 10, 30, color.NRGBA{0xff, 0, 0, 0xff}, fmt.Sprintf("Sector: %v[%v]", reflect.TypeOf(ps.Player.Sector), ps.Player.Sector.GetBase().ID))
	y := 0
	for y < 20 && renderer.DebugNotices.Length() > 0 {
		msg := renderer.DebugNotices.Pop().(string)
		mainFont.Draw(win, 10, 40+float64(y)*10, color.NRGBA{0xff, 0, 0, 0xff}, msg)
		y++
	}
	win.Update()
}

func run() {
	//debug.SetGCPercent(-1)
	//debug.SetMemoryLimit(1024 * 1024 * 1024)
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	w := 640
	h := 360
	cfg := pixelgl.WindowConfig{
		Title:     "Foom",
		Bounds:    pixel.R(0, 0, 3840, 2160),
		VSync:     true,
		Resizable: true,
		//Undecorated: true,
		//Monitor:     pixelgl.PrimaryMonitor(),
	}
	var err error
	win, err = pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.SetSmooth(false)

	canvas = pixelgl.NewCanvas(pixel.R(0, 0, float64(w), float64(h)))
	buffer = image.NewRGBA(image.Rect(0, 0, w, h))
	renderer = render.NewRenderer()
	renderer.ScreenWidth = w
	renderer.ScreenHeight = h
	renderer.Initialize()

	/*gameMap = logic.NewMapService(&core.Map{})
	gameMap.Initialize()
	gameMap.CreateTest()*/

	sim = core.NewSimulation()
	sim.Integrate = integrateGame
	sim.Render = renderGame
	gameMap = controllers.LoadMap("data/worlds/hall.json")
	gameMap.Attach(sim)
	ps = entity.NewPlayerController(gameMap.Player.(*entities.Player))
	ps.Collide()
	renderer.Map = gameMap.Map

	mainFont, _ = render.NewFont("/Library/Fonts/Courier New.ttf", 24)

	for !win.Closed() {
		sim.Step()
		//runtime.GC()
	}
}

func main() {
	flag.Parse()

	pixelgl.Run(run)
}
