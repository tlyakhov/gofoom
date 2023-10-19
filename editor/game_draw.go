package main

import (
	"fmt"
	"reflect"
	"tlyakhov/gofoom/entities"
	"tlyakhov/gofoom/logic/entity"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
)

func DrawGame(da *gtk.DrawingArea, cr *cairo.Context) {
	w := 640
	h := 360

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
	daw := da.GetAllocatedWidth()
	dah := da.GetAllocatedHeight()
	cr.Scale(float64(daw)/float64(w), float64(dah)/float64(h))
	cr.SetSourceSurface(editor.GameViewSurface, 0, 0)
	cr.Paint()

	ps := entity.NewPlayerService(editor.World.Player.(*entities.Player))

	cr.SetSourceRGB(1, 0, 1)
	cr.SelectFontFace("Courier", cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
	cr.SetFontSize(13)
	cr.MoveTo(10, 10)
	cr.ShowText(fmt.Sprintf("FPS: %.1f", editor.World.Sim().FPS))
	cr.MoveTo(10, 20)
	cr.ShowText(fmt.Sprintf("Health: %.1f", ps.Player.Health))
	cr.MoveTo(10, 30)
	cr.ShowText(fmt.Sprintf("Sector: %v[%v]", reflect.TypeOf(ps.Player.Sector), ps.Player.Sector.GetBase().ID))
}
