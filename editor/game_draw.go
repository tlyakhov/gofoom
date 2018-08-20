package main

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
)

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
