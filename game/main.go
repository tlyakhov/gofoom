// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"image"
	"log"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"

	"tlyakhov/gofoom/components/audio"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/ui"

	_ "image/png"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/render"
	_ "tlyakhov/gofoom/scripting_symbols"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

var cpuProfile = flag.String("cpuprofile", "", "Write CPU profile to file")
var memProfile = flag.String("memprofile", "", "Write Memory profile to file")
var win *opengl.Window
var renderer *render.Renderer
var canvas *opengl.Canvas
var buffer *image.RGBA
var inMenu = true

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
	} else if renderer.Player != nil {
		gameInput()
		ecs.ActAllControllers(ecs.ControllerFrame)
	}
	win.UpdateInput()
}

func renderGame() {
	winw := win.Bounds().W()
	winh := win.Bounds().H()

	w := 640
	h := 360
	if winw/winh < float64(w)/float64(h) {
		w = int(winw * float64(h) / winh)
	} else {
		h = int(winh * float64(w) / winw)
	}
	if w != renderer.ScreenWidth || h != renderer.ScreenHeight {
		log.Printf("New game canvas size: %vx%v", w, h)
		canvas = opengl.NewCanvas(pixel.R(0, 0, float64(w), float64(h)))
		buffer = image.NewRGBA(image.Rect(0, 0, w, h))
		renderer.ScreenWidth = w
		renderer.ScreenHeight = h
		renderer.Initialize()
		gameUI.Initialize()
	}

	renderer.Render()
	renderer.DebugInfo()
	if inMenu {
		gameUI.Render()
	}
	renderer.ApplyBuffer(buffer.Pix)

	canvas.SetPixels(buffer.Pix)

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
	cfg := opengl.WindowConfig{
		Title:     "Foom",
		Bounds:    pixel.R(0, 0, 1920, 1080),
		VSync:     false,
		Resizable: true,
		Maximized: true,
		//Undecorated: true,
		//Monitor:     opengl.PrimaryMonitor(),
	}
	var err error
	win, err = opengl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.SetSmooth(false)
	win.SetCursorDisabled()

	ecs.Initialize()
	ecs.Simulation.Integrate = integrateGame
	ecs.Simulation.Render = renderGame
	audio.Mixer.Initialize()
	defer audio.Mixer.Close()

	// Debug
	if false {
		controllers.CreateTestWorld3()
		// ecs.Save("bin/exported_test.yaml")
	} else if err = ecs.Load("data/worlds/pursuer-test.yaml"); err != nil {
		log.Printf("Error loading world %v", err)
		return
	}
	controllers.RespawnAll()
	controllers.CreateFont("data/fonts/vga-font-8x8.png", "Default Font")

	renderer = render.NewRenderer()
	gameUI = &ui.UI{Renderer: renderer}
	gameUI.OnChanged = onWidgetChanged
	gameUI.Initialize()
	initializeMenus()
	ui.LoadSettings(constants.UserSettings, uiPageMain, uiPageSettings, uiPageKeyBindings)

	for !win.Closed() {
		ecs.Simulation.Step()
	}

	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		// Lookup("allocs") creates a profile similar to go test -memprofile.
		// Alternatively, use Lookup("heap") for a profile
		// that has inuse_space as the default index.
		if err := pprof.Lookup("heap").WriteTo(f, 0); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

func main() {
	flag.Parse()

	opengl.Run(run)
}
