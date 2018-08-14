package main

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
	"github.com/tlyakhov/gofoom/concepts"
)

func DrawHandle(cr *cairo.Context, v concepts.Vector2) {
	v = editor.WorldToScreen(v)
	v1 := editor.ScreenToWorld(v.Sub(concepts.Vector2{3, 3}))
	v2 := editor.ScreenToWorld(v.Add(concepts.Vector2{3, 3}))
	cr.Rectangle(v1.X, v1.Y, v2.X-v1.X, v2.Y-v1.Y)
	cr.Stroke()
}

func DrawGame(da *gtk.DrawingArea, cr *cairo.Context) {
	w := da.GetAllocatedWidth()
	h := da.GetAllocatedHeight()

	if editor.Renderer == nil || editor.Renderer.ScreenWidth != w || editor.Renderer.ScreenHeight != h {
		editor.GameView(w, h)
	}
	editor.Renderer.Render(editor.GameViewBuffer)
	// Unfortunately our image surface is BGRA instead of RGBA. Let's swap:
	for i := 0; i < w*h*4; i += 4 {
		tmp := editor.GameViewBuffer[i]
		editor.GameViewBuffer[i] = editor.GameViewBuffer[i+2]
		editor.GameViewBuffer[i+2] = tmp
	}
	editor.GameViewSurface.MarkDirty()
	cr.SetSourceSurface(editor.GameViewSurface, 0, 0)
	cr.Paint()
}
