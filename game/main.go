// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"

	"tlyakhov/gofoom/archetypes"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/inventory"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/ecs"
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
var memProfile = flag.String("memprofile", "", "Write Memory profile to file")
var win *opengl.Window
var renderer *render.Renderer
var canvas *opengl.Canvas
var buffer *image.RGBA
var inMenu = true

// TODO: unify this with editor, and also add ability to customize keybinds.
// TODO: Mouse look?
func gameInput() {
	playerMobile := core.GetMobile(renderer.Player.Entity)

	if win.Pressed(pixel.KeyW) {
		controllers.MovePlayer(renderer.Player.Entity, renderer.PlayerBody.Angle.Now)
	}
	if win.Pressed(pixel.KeyS) {
		controllers.MovePlayer(renderer.Player.Entity, renderer.PlayerBody.Angle.Now+180.0)
	}
	if win.Pressed(pixel.KeyE) {
		controllers.MovePlayer(renderer.Player.Entity, renderer.PlayerBody.Angle.Now+90.0)
	}
	if win.Pressed(pixel.KeyQ) {
		controllers.MovePlayer(renderer.Player.Entity, renderer.PlayerBody.Angle.Now+270.0)
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
		if renderer.Carrier.SelectedWeapon != 0 {
			if w := inventory.GetWeapon(renderer.Carrier.SelectedWeapon); w != nil {
				w.Intent = inventory.WeaponFire
			}
		}
	}
	// renderer.Player.ShearZ = (win.MousePosition().Y - win.Bounds().H()*0.5) * 0.8
	renderer.Player.ActionPressed = (win.JustPressed(pixel.MouseButton2) || win.Repeated(pixel.MouseButton2))

	if win.Pressed(pixel.KeySpace) {

		if behaviors.GetUnderwater(renderer.PlayerBody.SectorEntity) != nil {
			playerMobile.Force[2] += constants.PlayerSwimStrength
		} else if renderer.PlayerBody.OnGround {
			playerMobile.Force[2] += constants.PlayerJumpForce
			renderer.PlayerBody.OnGround = false
		}
	}
	if win.Pressed(pixel.KeyC) {
		if behaviors.GetUnderwater(renderer.PlayerBody.SectorEntity) != nil {
			playerMobile.Force[2] -= constants.PlayerSwimStrength
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
	} else if renderer.Player != nil {
		if gunSlot != nil && gunSlot.Count.Now > 0 {
			renderer.Carrier.SelectedWeapon = gunSlot.Entity
		}
		gameInput()
		ecs.ActAllControllers(ecs.ControllerAlways)
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

	ecs.Initialize()
	for _, meta := range ecs.Types().Controllers {
		fmt.Printf("%v, priority %v\n", meta.Type.String(), meta.Priority)
	}
	ecs.Simulation.Integrate = integrateGame
	ecs.Simulation.Render = renderGame
	// Debug
	if false {
		controllers.CreateTestWorld3()
		// ecs.Save("bin/exported_test.yaml")
	} else if err = ecs.Load("data/worlds/hall.yaml"); err != nil {
		log.Printf("Error loading world %v", err)
		return
	}
	validateSpawn()
	controllers.Respawn(true)
	archetypes.CreateFont("data/vga-font-8x8.png", "Default Font")

	canvas = opengl.NewCanvas(pixel.R(0, 0, float64(w), float64(h)))
	buffer = image.NewRGBA(image.Rect(0, 0, w, h))
	renderer = render.NewRenderer()
	renderer.ScreenWidth = w
	renderer.ScreenHeight = h
	renderer.Initialize()
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
