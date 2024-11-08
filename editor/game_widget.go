// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"image"
	"log"
	"tlyakhov/gofoom/containers"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/fogleman/gg"
)

// Declare conformity with interfaces
var _ fyne.Focusable = (*GameWidget)(nil)
var _ fyne.Widget = (*GameWidget)(nil)
var _ desktop.Mouseable = (*GameWidget)(nil)
var _ desktop.Keyable = (*GameWidget)(nil)

type GameWidget struct {
	widget.BaseWidget

	Raster  *canvas.Raster
	Context *gg.Context
	Surface *image.RGBA
	KeyMap  containers.Set[fyne.KeyName]
}

func NewGameWidget() *GameWidget {
	g := &GameWidget{
		KeyMap: make(containers.Set[fyne.KeyName]),
	}
	g.ExtendBaseWidget(g)
	g.Raster = canvas.NewRaster(g.generateRaster)
	return g
}

func (g *GameWidget) generateRaster(requestedWidth, requestedHeight int) image.Image {
	editor.Lock.Lock()
	defer editor.Lock.Unlock()

	w := 640
	h := w * requestedHeight / requestedWidth
	if g.Surface != nil && g.Context != nil &&
		editor.Renderer.ScreenWidth == w && editor.Renderer.ScreenHeight == h {
		return g.Surface
	}
	editor.ResizeRenderer(w, h)
	g.Surface = image.NewRGBA(image.Rect(0, 0, w, h))
	g.Context = gg.NewContext(w, h)
	return g.Surface
}

// MinSize returns the size that this widget should not shrink below
func (g *GameWidget) MinSize() fyne.Size {
	g.ExtendBaseWidget(g)
	return g.BaseWidget.MinSize()
}

func (g *GameWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(g.Raster)
}

func (g *GameWidget) Draw() {
	if g.Context == nil || editor.Renderer == nil {
		return
	}

	editor.Lock.Lock()
	pixels := g.Context.Image().(*image.RGBA).Pix
	editor.Renderer.Render()
	editor.Renderer.DebugInfo()
	editor.Renderer.ApplyBuffer(pixels)
	editor.Lock.Unlock()

	copy(g.Surface.Pix, pixels)
	g.Raster.Refresh()
}

func (g *GameWidget) KeyDown(evt *fyne.KeyEvent) {
	editor.Lock.Lock()
	defer editor.Lock.Unlock()
	g.KeyMap[evt.Name] = struct{}{}
}
func (g *GameWidget) KeyUp(evt *fyne.KeyEvent) {
	editor.Lock.Lock()
	defer editor.Lock.Unlock()
	delete(g.KeyMap, evt.Name)
}

func (g *GameWidget) FocusLost()       {}
func (g *GameWidget) FocusGained()     {}
func (g *GameWidget) TypedRune(r rune) {}
func (g *GameWidget) TypedKey(evt *fyne.KeyEvent) {
	for _, action := range editor.MenuActions {
		if action.NoModifier && evt.Name == action.Shortcut.KeyName {
			action.Menu.Action()
		}
	}
}

func (g *GameWidget) MouseDown(evt *desktop.MouseEvent) {
	g.requestFocus()

	// TODO: make this more granular, and also support bodies
	if evt.Button == desktop.MouseButtonSecondary {
		daw := g.Raster.Size().Width
		dah := g.Raster.Size().Height
		rw := editor.Renderer.ScreenWidth
		rh := editor.Renderer.ScreenHeight
		x := float64(evt.Position.X) * float64(rw) / float64(daw)
		y := float64(evt.Position.Y) * float64(rh) / float64(dah)
		picked := editor.Renderer.Pick(int(x), int(y))
		if evt.Modifier&fyne.KeyModifierShift != 0 {
			editor.SelectedObjects.Add(picked...)
			editor.SetSelection(true, editor.SelectedObjects)
		} else if evt.Modifier&fyne.KeyModifierSuper != 0 {
			log.Printf("Subtracting game picking unimplemented")
		} else {
			editor.SelectObjects(true, picked...)
		}
	}
}
func (g *GameWidget) MouseUp(evt *desktop.MouseEvent) {}

func (g *GameWidget) requestFocus() {
	if c := fyne.CurrentApp().Driver().CanvasForObject(g); c != nil {
		c.Focus(g)
	}
}
