package main

import (
	"image"

	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/math"

	"github.com/tlyakhov/gofoom/engine"

	// "math"
	// "math/rand"
	// "os"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Foom",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
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

	buffer := image.NewRGBA(image.Rect(0, 0, 1024, 768))

	gameMap := engine.Map{}

	renderer := engine.NewRenderer()

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		win.SetClosed(win.JustPressed(pixelgl.KeyEscape))

		if win.JustPressed(pixelgl.MouseButtonLeft) {
		}
		if win.Pressed(pixelgl.KeyW) {
			gameMap.Player.Move(gameMap.Player.Angle, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyS) {
			gameMap.Player.Move(gameMap.Player.Angle+180.0, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyQ) {
			gameMap.Player.Move(gameMap.Player.Angle+270.0, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyE) {
			gameMap.Player.Move(gameMap.Player.Angle+90.0, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyQ) {
			gameMap.Player.Move(gameMap.Player.Angle+270.0, dt, 1.0)
		}
		if win.Pressed(pixelgl.KeyA) {
			gameMap.Player.Angle -= constants.PlayerTurnSpeed * dt / 30.0
			gameMap.Player.Angle = math.NormalizeAngle(gameMap.Player.Angle)
		}
		if win.Pressed(pixelgl.KeyD) {
			gameMap.Player.Angle += constants.PlayerTurnSpeed * dt / 30.0
			gameMap.Player.Angle = math.NormalizeAngle(gameMap.Player.Angle)
		}

		renderer.Render(buffer.Pix)

		canvas.SetPixels(buffer.Pix)
		canvas.Draw(win, mat)
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
