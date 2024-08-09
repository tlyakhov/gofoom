// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"image"
	"log"
	"os"
	"runtime/pprof"

	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/ui"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/render"
	_ "tlyakhov/gofoom/scripting_symbols"

	// "math"
	// "math/rand"
	// "os"

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
var inMenu = true

func gameInput() {
	if renderer.Player == nil {
		return
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
	if win.JustPressed(pixel.MouseButton1) || win.Repeated(pixel.MouseButton1) {
		if w := behaviors.WeaponInstantFromDb(renderer.DB, renderer.Player.Entity); w != nil {
			w.FireNextFrame = true
		}
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
	if win.JustReleased(pixel.KeyEscape) {
		if inMenu && gameUI.Page != nil && gameUI.Page.Parent != nil {
			gameUI.SetPage(gameUI.Page.Parent)
		} else {
			inMenu = !inMenu
			if !inMenu {
				gameUI.SetPage(nil)
			}
		}
	}

	if inMenu {
		menuInput()
	} else {
		gameInput()
		db.ActAllControllers(concepts.ControllerAlways)
	}
	win.UpdateInput()
}

func renderGame() {
	renderer.Render()
	renderer.DebugInfo()
	if inMenu {
		gameUI.Render()
	}
	renderer.ApplyBuffer(buffer.Pix)

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
	win.SwapBuffers()
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
	gameUI = &ui.UI{Renderer: renderer}
	gameUI.OnChanged = saveSettings
	gameUI.Initialize()
	initializeMenus()
	ui.LoadSettings(constants.UserSettings, uiPageMain, uiPageOptions, uiPageKeyBindings)

	for !win.Closed() {
		db.Simulation.Step()
	}
}

func main() {
	flag.Parse()

	opengl.Run(run)
}
