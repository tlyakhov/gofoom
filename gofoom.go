package main

import (
	"fmt"
	"image"

	"github.com/tlyakhov/gofoom/engine/mapping"

	// "math"
	// "math/rand"
	// "os"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"

	"github.com/tlyakhov/gofoom/engine"
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

	m := mapping.Map{}
	m.GenerateID()
	fmt.Println(m.ID)
	return

	renderer := engine.NewRenderer()

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		_ = dt

		win.SetClosed(win.JustPressed(pixelgl.KeyEscape))

		if win.JustPressed(pixelgl.MouseButtonLeft) {
		}
		if win.Pressed(pixelgl.KeyLeft) {
		}
		if win.Pressed(pixelgl.KeyRight) {
		}
		if win.Pressed(pixelgl.KeyDown) {
		}
		if win.Pressed(pixelgl.KeyUp) {
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
