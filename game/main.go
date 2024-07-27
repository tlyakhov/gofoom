// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"runtime/pprof"

	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/controllers"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/render"
	_ "tlyakhov/gofoom/scripting_symbols"

	// "math"
	// "math/rand"
	// "os"

	"image/color"
	_ "image/png"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

var cpuProfile = flag.String("cpuprofile", "", "Write CPU profile to file")
var win *opengl.Window
var db *concepts.EntityComponentDB
var renderer *render.Renderer
var canvas *opengl.Canvas
var buffer *image.RGBA
var mainFont *render.Font

func processInput() {
	if renderer.Player == nil {
		return
	}
	win.SetClosed(win.JustPressed(pixel.KeyEscape))

	if win.JustPressed(pixel.MouseButtonLeft) {
	}
	if win.Pressed(pixel.KeyW) {
		controllers.MovePlayer(renderer.PlayerBody, renderer.PlayerBody.Angle.Now, false)
	}
	if win.Pressed(pixel.KeyS) {
		controllers.MovePlayer(renderer.PlayerBody, renderer.PlayerBody.Angle.Now+180.0, false)
	}
	if win.Pressed(pixel.KeyE) {
		controllers.MovePlayer(renderer.PlayerBody, renderer.PlayerBody.Angle.Now+90.0, false)
	}
	if win.Pressed(pixel.KeyQ) {
		controllers.MovePlayer(renderer.PlayerBody, renderer.PlayerBody.Angle.Now+270.0, false)
	}
	if win.Pressed(pixel.KeyA) {
		renderer.PlayerBody.Angle.Now -= constants.PlayerTurnSpeed * constants.TimeStepS
		renderer.PlayerBody.Angle.Now = concepts.NormalizeAngle(renderer.PlayerBody.Angle.Now)
	}
	if win.Pressed(pixel.KeyD) {
		renderer.PlayerBody.Angle.Now += constants.PlayerTurnSpeed * constants.TimeStepS
		renderer.PlayerBody.Angle.Now = concepts.NormalizeAngle(renderer.PlayerBody.Angle.Now)
	}
	if win.Pressed(pixel.KeySpace) {

		if behaviors.UnderwaterFromDb(renderer.DB, renderer.PlayerBody.SectorEntity) != nil {
			renderer.PlayerBody.Force[2] += constants.PlayerSwimStrength
		} else if renderer.PlayerBody.OnGround {
			renderer.PlayerBody.Force[2] += constants.PlayerJumpForce
			renderer.PlayerBody.OnGround = false
		}
	}
	if win.Pressed(pixel.KeyC) {
		if behaviors.UnderwaterFromDb(renderer.DB, renderer.PlayerBody.SectorEntity) != nil {
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
	db.ActAllControllers(concepts.ControllerAlways)
}

func renderGame() {
	playerAlive := behaviors.AliveFromDb(renderer.DB, renderer.PlayerBody.Entity)
	// player := bodies.PlayerFromDb(&gameMap.Player)

	renderer.Render(buffer.Pix)
	canvas.SetPixels(buffer.Pix)
	winw := win.Bounds().W()
	winh := win.Bounds().H()
	mat := pixel.IM
	mat = mat.ScaledXY(pixel.ZV, pixel.Vec{X: winw / float64(renderer.ScreenWidth), Y: winh / float64(renderer.ScreenHeight)})
	win.SetMatrix(mat)
	mat = pixel.IM
	mat = mat.ScaledXY(pixel.ZV, pixel.Vec{X: 1, Y: -1})
	mat = mat.Moved(pixel.Vec{X: float64(renderer.ScreenWidth / 2), Y: float64(renderer.ScreenHeight / 2)})
	canvas.Draw(win, mat)

	screen := renderer.WorldToScreen(&concepts.Vector3{-320, 30, -32})
	if screen != nil {
		mainFont.Draw(win, screen[0], float64(renderer.ScreenHeight)-screen[1], color.NRGBA{0xff, 0, 0, 0xff}, "test")
	}
	mainFont.Draw(win, 10, 10, color.NRGBA{0xff, 0, 0, 0xff}, fmt.Sprintf("FPS: %.1f, Light cache: %v", renderer.DB.Simulation.FPS, renderer.SectorLastRendered.Size()))
	mainFont.Draw(win, 10, 20, color.NRGBA{0xff, 0, 0, 0xff}, fmt.Sprintf("Health: %.1f", playerAlive.Health))
	hits := renderer.ICacheHits.Load()
	misses := renderer.ICacheMisses.Load()
	mainFont.Draw(win, 10, 30, color.NRGBA{0xff, 0, 0, 0xff}, fmt.Sprintf("ICache hit percentage: %.1f, %v, %v", float64(hits)*100.0/float64(hits+misses), hits, misses))
	if renderer.PlayerBody.SectorEntity != 0 {
		entity := renderer.PlayerBody.SectorEntity
		s := 0
		//		core.SectorFromDb(ref).Lightmap.Range(func(k uint64, v concepts.Vector4) bool { s++; return true })
		mainFont.Draw(win, 10, 40, color.NRGBA{0xff, 0, 0, 0xff}, fmt.Sprintf("Sector: %v, LM:%v", entity.String(renderer.DB), s))
	}
	y := 0
	for y < 20 && renderer.DebugNotices.Length() > 0 {
		msg := renderer.DebugNotices.Pop().(string)
		mainFont.Draw(win, 10, 50+float64(y)*10, color.NRGBA{0xff, 0, 0, 0xff}, msg)
		y++
	}
	win.Update()
}

func run() {
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
	cfg := opengl.WindowConfig{
		Title:     "Foom",
		Bounds:    pixel.R(0, 0, 1920, 1080),
		VSync:     false,
		Resizable: true,
		//Undecorated: true,
		//Monitor:     opengl.PrimaryMonitor(),
	}
	var err error
	win, err = opengl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.SetSmooth(false)

	db = concepts.NewEntityComponentDB()
	db.Simulation.Integrate = integrateGame
	db.Simulation.Render = renderGame
	//controllers.CreateTestWorld(db)
	//db.Save("data/worlds/exported_test.json")
	if err = db.Load("data/worlds/hall.json"); err != nil {
		log.Printf("Error loading world %v", err)
		return
	}
	canvas = opengl.NewCanvas(pixel.R(0, 0, float64(w), float64(h)))
	buffer = image.NewRGBA(image.Rect(0, 0, w, h))
	renderer = render.NewRenderer(db)
	renderer.ScreenWidth = w
	renderer.ScreenHeight = h
	renderer.Initialize()

	mainFont, _ = render.NewFont("/Library/Fonts/Courier New.ttf", 24)

	for !win.Closed() {
		db.Simulation.Step()
	}
}

func main() {
	flag.Parse()

	opengl.Run(run)
}
