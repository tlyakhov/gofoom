package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"runtime/pprof"

	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/sectors"
	"tlyakhov/gofoom/controllers"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
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
var db *concepts.EntityComponentDB
var renderer *render.Renderer
var gameMap *core.Spawn
var canvas *pixelgl.Canvas
var buffer *image.RGBA
var mainFont *render.Font

func processInput() {
	win.SetClosed(win.JustPressed(pixelgl.KeyEscape))

	if win.JustPressed(pixelgl.MouseButtonLeft) {
	}
	if win.Pressed(pixelgl.KeyW) {
		controllers.MovePlayer(renderer.PlayerBody, renderer.PlayerBody.Angle)
	}
	if win.Pressed(pixelgl.KeyS) {
		controllers.MovePlayer(renderer.PlayerBody, renderer.PlayerBody.Angle+180.0)
	}
	if win.Pressed(pixelgl.KeyE) {
		controllers.MovePlayer(renderer.PlayerBody, renderer.PlayerBody.Angle+90.0)
	}
	if win.Pressed(pixelgl.KeyQ) {
		controllers.MovePlayer(renderer.PlayerBody, renderer.PlayerBody.Angle+270.0)
	}
	if win.Pressed(pixelgl.KeyA) {
		renderer.PlayerBody.Angle -= constants.PlayerTurnSpeed * constants.TimeStepS
		renderer.PlayerBody.Angle = concepts.NormalizeAngle(renderer.PlayerBody.Angle)
	}
	if win.Pressed(pixelgl.KeyD) {
		renderer.PlayerBody.Angle += constants.PlayerTurnSpeed * constants.TimeStepS
		renderer.PlayerBody.Angle = concepts.NormalizeAngle(renderer.PlayerBody.Angle)
	}
	if win.Pressed(pixelgl.KeySpace) {

		if renderer.PlayerBody.SectorEntityRef.Component(sectors.UnderwaterComponentIndex) != nil {
			renderer.PlayerBody.Force[2] += constants.PlayerSwimStrength
		} else if renderer.PlayerBody.OnGround {
			renderer.PlayerBody.Force[2] += constants.PlayerJumpForce
			renderer.PlayerBody.OnGround = false
		}
	}
	if win.Pressed(pixelgl.KeyC) {
		if renderer.PlayerBody.SectorEntityRef.Component(sectors.UnderwaterComponentIndex) != nil {
			renderer.PlayerBody.Force[2] -= constants.PlayerSwimStrength
		} else {
			renderer.Player.Crouching = true
		}
	} else {
		renderer.Player.Crouching = false
	}
}

func integrateGame() {
	processInput()
	db.NewControllerSet().ActGlobal(concepts.ControllerAlways)
}

func renderGame() {
	playerAlive := behaviors.AliveFromDb(renderer.PlayerBody.EntityRef)
	// player := bodies.PlayerFromDb(&gameMap.Player)

	renderer.Render(buffer.Pix)
	canvas.SetPixels(buffer.Pix)
	winw := win.Bounds().W()
	winh := win.Bounds().H()
	mat := pixel.IM.ScaledXY(pixel.Vec{X: 0, Y: 0}, pixel.Vec{X: winw / float64(renderer.ScreenWidth), Y: -winh / float64(renderer.ScreenHeight)}).Moved(win.Bounds().Center())
	canvas.Draw(win, mat)
	mainFont.Draw(win, 10, 10, color.NRGBA{0xff, 0, 0, 0xff}, fmt.Sprintf("FPS: %.1f", db.Simulation.FPS))
	mainFont.Draw(win, 10, 20, color.NRGBA{0xff, 0, 0, 0xff}, fmt.Sprintf("Health: %.1f", playerAlive.Health))
	if !renderer.PlayerBody.SectorEntityRef.Nil() {
		mainFont.Draw(win, 10, 30, color.NRGBA{0xff, 0, 0, 0xff}, fmt.Sprintf("Sector: %v", renderer.PlayerBody.SectorEntityRef.String()))
	}
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

		/*	// Start tracing
			traceFile, err := os.Create("trace.out")
			if err != nil {
				panic(err)
			}
			defer traceFile.Close()

			if err := trace.Start(traceFile); err != nil {
				panic(err)
			}
			defer trace.Stop()*/
	}

	w := 640
	h := 360
	cfg := pixelgl.WindowConfig{
		Title:     "Foom",
		Bounds:    pixel.R(0, 0, 1920, 1080),
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

	db = concepts.NewEntityComponentDB()
	db.Simulation.Integrate = integrateGame
	db.Simulation.Render = renderGame

	canvas = pixelgl.NewCanvas(pixel.R(0, 0, float64(w), float64(h)))
	buffer = image.NewRGBA(image.Rect(0, 0, w, h))
	renderer = render.NewRenderer(db)
	renderer.ScreenWidth = w
	renderer.ScreenHeight = h
	renderer.Initialize()

	//controllers.CreateTestWorld2(db)
	//db.Save("data/worlds/exported_test.json")
	if err = db.Load("data/worlds/hall.json"); err != nil {
		log.Printf("Error loading world %v", err)
		return
	}
	renderer.DB = db

	mainFont, _ = render.NewFont("/Library/Fonts/Courier New.ttf", 24)

	for !win.Closed() {
		db.Simulation.Step()
		//runtime.GC()
	}
}

func main() {
	flag.Parse()

	pixelgl.Run(run)
}
