package main

import (
	"encoding/json"
	"flag"
	"image"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/logic"
	"github.com/tlyakhov/gofoom/mapping"
	"github.com/tlyakhov/gofoom/registry"
	"github.com/tlyakhov/gofoom/render"

	// "math"
	// "math/rand"
	// "os"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var cpuProfile = flag.String("cpuprofile", "", "Write CPU profile to file")

func loadMap(filename string) *mapping.Map {
	fileContents, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}
	var parsed interface{}
	err = json.Unmarshal(fileContents, &parsed)
	m := &mapping.Map{}
	m.Initialize()
	m.Deserialize(parsed.(map[string]interface{}))
	return m
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
	cfg := pixelgl.WindowConfig{
		Title:  "Foom",
		Bounds: pixel.R(0, 0, 800, 600),
		VSync:  false,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.SetSmooth(false)

	canvas := pixelgl.NewCanvas(win.Bounds())

	var (
		mat = pixel.IM.Moved(win.Bounds().Center())
	)

	buffer := image.NewRGBA(image.Rect(0, 0, 800, 600))
	renderer := render.NewRenderer()
	renderer.ScreenWidth = 800
	renderer.ScreenHeight = 600
	renderer.WorkerWidth = 800
	renderer.Initialize()
	gameMap := loadMap("data/testMap.json")
	player := registry.Translate(gameMap.Player, "logic").(*logic.Player)
	registry.Translate(&player.Entity, "logic").(*logic.Entity).Collide()
	renderer.Map = gameMap

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds() * 1000
		last = time.Now()

		win.SetClosed(win.JustPressed(pixelgl.KeyEscape))

		if win.JustPressed(pixelgl.MouseButtonLeft) {
		}
		if win.Pressed(pixelgl.KeyW) {
			player.Move(gameMap.Player.Angle, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyS) {
			player.Move(gameMap.Player.Angle+180.0, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyQ) {
			player.Move(gameMap.Player.Angle+270.0, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyE) {
			player.Move(gameMap.Player.Angle+90.0, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyQ) {
			player.Move(gameMap.Player.Angle+270.0, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyA) {
			player.Angle -= constants.PlayerTurnSpeed * dt / 30.0
			player.Angle = concepts.NormalizeAngle(player.Angle)
		}
		if win.Pressed(pixelgl.KeyD) {
			player.Angle += constants.PlayerTurnSpeed * dt / 30.0
			player.Angle = concepts.NormalizeAngle(player.Angle)
		}

		renderer.Render(buffer.Pix)
		registry.Translate(gameMap, "logic").(*logic.Map).Frame(dt)

		canvas.SetPixels(buffer.Pix)
		canvas.Draw(win, mat)
		win.Update()
	}
}

func main() {
	flag.Parse()

	pixelgl.Run(run)
}
