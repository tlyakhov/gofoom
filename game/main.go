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
var ps *entity.PlayerController
var sim *core.Simulation
var renderer *render.Renderer
var gameMap *controllers.MapController
var canvas *pixelgl.Canvas
var buffer *image.RGBA
var mainFont *render.Font

func processInput() {
	player := gameMap.Player.(*entities.Player)
	win.SetClosed(win.JustPressed(pixelgl.KeyEscape))

	if win.JustPressed(pixelgl.MouseButtonLeft) {
	}
	if win.Pressed(pixelgl.KeyW) {
		ps.Move(player.Angle)
	}
	if win.Pressed(pixelgl.KeyS) {
		ps.Move(player.Angle + 180.0)
	}
	if win.Pressed(pixelgl.KeyE) {
		ps.Move(player.Angle + 90.0)
	}
	if win.Pressed(pixelgl.KeyQ) {
		ps.Move(player.Angle + 270.0)
	}
	if win.Pressed(pixelgl.KeyA) {
		player.Angle -= constants.PlayerTurnSpeed * constants.TimeStep
		player.Angle = concepts.NormalizeAngle(player.Angle)
	}
	if win.Pressed(pixelgl.KeyD) {
		player.Angle += constants.PlayerTurnSpeed * constants.TimeStep
		player.Angle = concepts.NormalizeAngle(player.Angle)
	}
	if win.Pressed(pixelgl.KeySpace) {
		if _, ok := player.Sector.(*sectors.Underwater); ok {
			player.Vel.Now[2] += constants.PlayerSwimStrength * constants.TimeStep
		} else if player.OnGround {
			player.Vel.Now[2] += constants.PlayerJumpForce * constants.TimeStep
			player.OnGround = false
		}
	}
	if win.Pressed(pixelgl.KeyC) {
		if _, ok := player.Sector.(*sectors.Underwater); ok {
			player.Vel.Now[2] -= constants.PlayerSwimStrength * constants.TimeStep
		} else {
			player.Crouching = true
		}
	} else {
		player.Crouching = false
	}
}

func integrateGame() {
	processInput()
	gameMap.Frame()
}

func renderGame() {
	player := gameMap.Player.(*entities.Player)

	renderer.Render(buffer.Pix)
	canvas.SetPixels(buffer.Pix)
	winw := win.Bounds().W()
	winh := win.Bounds().H()
	mat := pixel.IM.ScaledXY(pixel.Vec{X: 0, Y: 0}, pixel.Vec{X: winw / float64(renderer.ScreenWidth), Y: -winh / float64(renderer.ScreenHeight)}).Moved(win.Bounds().Center())
	canvas.Draw(win, mat)
	mainFont.Draw(win, 10, 10, color.NRGBA{0xff, 0, 0, 0xff}, fmt.Sprintf("FPS: %.1f", sim.FPS))
	mainFont.Draw(win, 10, 20, color.NRGBA{0xff, 0, 0, 0xff}, fmt.Sprintf("Health: %.1f", player.Health))
	mainFont.Draw(win, 10, 30, color.NRGBA{0xff, 0, 0, 0xff}, fmt.Sprintf("Sector: %v[%v]", reflect.TypeOf(player.Sector), player.Sector.GetBase().ID))
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
	gameMap, err = controllers.LoadMap("data/worlds/hall.json")
	if err != nil {
		log.Printf("Error loading world %v", err)
		return
	}
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
