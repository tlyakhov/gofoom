package main

import (
	"encoding/json"
	"flag"
	"image"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"

	"github.com/tlyakhov/gofoom/logic"
	"github.com/tlyakhov/gofoom/logic/entity"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	_ "github.com/tlyakhov/gofoom/logic/entity"
	_ "github.com/tlyakhov/gofoom/logic/provide"
	_ "github.com/tlyakhov/gofoom/logic/sector"
	"github.com/tlyakhov/gofoom/mapping"
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
		Title:     "Foom",
		Bounds:    pixel.R(0, 0, 1280, 720),
		VSync:     false,
		Resizable: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.SetSmooth(false)

	canvas := pixelgl.NewCanvas(win.Bounds())

	var (
		mat = pixel.IM.ScaledXY(pixel.Vec{0, 0}, pixel.Vec{1, -1}).Moved(win.Bounds().Center())
	)

	buffer := image.NewRGBA(image.Rect(0, 0, 1280, 720))
	renderer := render.NewRenderer()
	renderer.ScreenWidth = 1280
	renderer.ScreenHeight = 720
	renderer.WorkerWidth = 1280
	renderer.Initialize()
	gameMap := loadMap("data/classicMap.json")
	ps := entity.NewPlayerService(gameMap.Player)
	ps.Collide()
	renderer.Map = gameMap

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds() * 1000
		last = time.Now()

		win.SetClosed(win.JustPressed(pixelgl.KeyEscape))

		if win.JustPressed(pixelgl.MouseButtonLeft) {
		}
		if win.Pressed(pixelgl.KeyW) {
			ps.Move(gameMap.Player.Angle, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyS) {
			ps.Move(gameMap.Player.Angle+180.0, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyQ) {
			ps.Move(gameMap.Player.Angle+270.0, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyE) {
			ps.Move(gameMap.Player.Angle+90.0, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyQ) {
			ps.Move(gameMap.Player.Angle+270.0, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyA) {
			ps.Player.Angle -= constants.PlayerTurnSpeed * dt / 30.0
			ps.Player.Angle = concepts.NormalizeAngle(ps.Player.Angle)
		}
		if win.Pressed(pixelgl.KeyD) {
			ps.Player.Angle += constants.PlayerTurnSpeed * dt / 30.0
			ps.Player.Angle = concepts.NormalizeAngle(ps.Player.Angle)
		}
		if win.Pressed(pixelgl.KeySpace) {
			if ps.Standing {
				ps.Player.Vel.Z += constants.PlayerJumpStrength * dt / 30.0
				ps.Standing = false
			}
		}

		renderer.Render(buffer.Pix)
		logic.NewMapService(gameMap).Frame(dt)

		canvas.SetPixels(buffer.Pix)
		canvas.Draw(win, mat)
		win.Update()
	}
}

func main() {
	flag.Parse()

	pixelgl.Run(run)
}
