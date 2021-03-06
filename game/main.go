package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"reflect"
	"runtime/pprof"

	_ "github.com/tlyakhov/gofoom/behaviors"
	"github.com/tlyakhov/gofoom/entities"
	"github.com/tlyakhov/gofoom/logic"
	"github.com/tlyakhov/gofoom/logic/entity"
	"github.com/tlyakhov/gofoom/sectors"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	_ "github.com/tlyakhov/gofoom/logic/provide"
	_ "github.com/tlyakhov/gofoom/logic/sector"
	"github.com/tlyakhov/gofoom/render"

	// "math"
	// "math/rand"
	// "os"
	"time"

	"image/color"
	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var cpuProfile = flag.String("cpuprofile", "", "Write CPU profile to file")

func run() {
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	w := 1280
	h := 720
	cfg := pixelgl.WindowConfig{
		Title:     "Foom",
		Bounds:    pixel.R(0, 0, float64(w), float64(h)),
		VSync:     true,
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

	buffer := image.NewRGBA(image.Rect(0, 0, w, h))
	renderer := render.NewRenderer()
	renderer.ScreenWidth = w
	renderer.ScreenHeight = h
	renderer.Initialize()
	gameMap := logic.LoadMap("data/worlds/hall.json")
	ps := entity.NewPlayerService(gameMap.Player.(*entities.Player))
	ps.Collide()
	renderer.Map = gameMap.Map

	mainFont, _ := render.NewFont("/Library/Fonts/Courier New.ttf", 24)

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds() * 1000
		last = time.Now()

		win.SetClosed(win.JustPressed(pixelgl.KeyEscape))

		if win.JustPressed(pixelgl.MouseButtonLeft) {
		}
		if win.Pressed(pixelgl.KeyW) {
			ps.Move(ps.Player.Angle, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyS) {
			ps.Move(ps.Player.Angle+180.0, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyE) {
			ps.Move(ps.Player.Angle+90.0, dt, 0.5)
		}
		if win.Pressed(pixelgl.KeyQ) {
			ps.Move(ps.Player.Angle+270.0, dt, 0.5)
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
			if _, ok := ps.Player.Sector.(*sectors.Underwater); ok {
				ps.Player.Vel.Z += constants.PlayerSwimStrength * dt / 30.0
			} else if ps.Standing {
				ps.Player.Vel.Z += constants.PlayerJumpStrength * dt / 30.0
				ps.Standing = false
			}
		}
		if win.Pressed(pixelgl.KeyC) {
			if _, ok := ps.Player.Sector.(*sectors.Underwater); ok {
				ps.Player.Vel.Z -= constants.PlayerSwimStrength * dt / 30.0
			} else {
				ps.Crouching = true
			}
		} else {
			ps.Crouching = false
		}
		gameMap.Frame(dt)
		renderer.Render(buffer.Pix)

		canvas.SetPixels(buffer.Pix)
		canvas.Draw(win, mat)
		mainFont.Draw(win, 10, 10, color.NRGBA{0xff, 0, 0, 0xff}, fmt.Sprintf("FPS: %.1f", 1000.0/dt))
		mainFont.Draw(win, 10, 20, color.NRGBA{0xff, 0, 0, 0xff}, fmt.Sprintf("Health: %.1f", ps.Player.Health))
		mainFont.Draw(win, 10, 30, color.NRGBA{0xff, 0, 0, 0xff}, fmt.Sprintf("Sector: %v[%v]", reflect.TypeOf(ps.Player.Sector), ps.Player.Sector.GetBase().ID))
		win.Update()
	}
}

func main() {
	flag.Parse()

	pixelgl.Run(run)
}
