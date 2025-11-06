// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"image"
	"log"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"
	"tlyakhov/gofoom/render"

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

	lastPick *render.PickResult
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

func (g *GameWidget) renderLastPick() {
	if g.lastPick == nil {
		return
	}

	r := editor.Renderer
	ts := r.NewTextStyle()
	ts.Color[0] = 0.2
	ts.Color[1] = 0.2
	ts.Color[2] = 1
	ts.Color[3] = 0.5
	ts.Shadow = false
	ts.HAnchor = 0
	ts.VAnchor = 0

	scr := r.WorldToScreen(&g.lastPick.World)
	if scr == nil {
		return
	}
	text := fmt.Sprintf("%v", g.lastPick.World.StringHuman(2))
	var sector *core.Sector
	for _, s := range g.lastPick.Selection {
		if s.Sector != nil {
			sector = s.Sector
		}
	}
	if sector != nil {
		text += fmt.Sprintf(" (0x%x)", r.WorldToLightmapHash(sector, &g.lastPick.World, &g.lastPick.Normal))
	}
	r.Print(ts, int(scr[0]), int(scr[1]), text)
}

func (g *GameWidget) Draw() {
	r := editor.Renderer

	if g.Context == nil || r == nil {
		return
	}

	editor.Lock.Lock()
	editor.MapWidget.render()
	pixels := g.Context.Image().(*image.RGBA).Pix
	r.Render()
	r.DebugInfo()

	g.renderLastPick()

	r.ApplyBuffer(pixels)
	editor.Lock.Unlock()

	copy(g.Surface.Pix, pixels)
	fyne.Do(g.Raster.Refresh)
}

func (g *GameWidget) KeyDown(evt *fyne.KeyEvent) {
	editor.GameInputLock.Lock()
	defer editor.GameInputLock.Unlock()
	g.KeyMap[evt.Name] = struct{}{}
}
func (g *GameWidget) KeyUp(evt *fyne.KeyEvent) {
	editor.GameInputLock.Lock()
	defer editor.GameInputLock.Unlock()
	delete(g.KeyMap, evt.Name)
}

func (g *GameWidget) FocusLost()       {}
func (g *GameWidget) FocusGained()     {}
func (g *GameWidget) TypedRune(r rune) {}
func (g *GameWidget) TypedKey(evt *fyne.KeyEvent) {
	for _, action := range editor.MenuActions {
		if ks, ok := action.Shortcut.(*desktop.CustomShortcut); ok {
			if action.NoModifier && evt.Name == ks.KeyName {
				action.Menu.Action()
			}
		}
	}
}

func (g *GameWidget) TypedShortcut(s fyne.Shortcut) {
	switch s.ShortcutName() {
	case "Undo":
		editor.UndoOrRedo(false)
	case "Redo":
		editor.UndoOrRedo(true)
	case "Cut":
		editor.Act(&actions.Copy{Action: state.Action{IEditor: editor}, Cut: true})
	case "Copy":
		editor.Act(&actions.Copy{Action: state.Action{IEditor: editor}, Cut: false})
	case "Paste":
		editor.Act(&actions.Paste{Transform: actions.Transform{Action: state.Action{IEditor: editor}}})
	}
}

func (g *GameWidget) MouseDown(evt *desktop.MouseEvent) {
	g.requestFocus()

	daw := g.Raster.Size().Width
	dah := g.Raster.Size().Height
	rw := editor.Renderer.ScreenWidth
	rh := editor.Renderer.ScreenHeight
	x := float64(evt.Position.X) * float64(rw) / float64(daw)
	y := float64(evt.Position.Y) * float64(rh) / float64(dah)

	if evt.Button == desktop.MouseButtonTertiary {
		g.lastPick = editor.Renderer.Pick(int(x), int(y))

	}

	// TODO: make this more granular, and also support bodies
	if evt.Button == desktop.MouseButtonSecondary {
		g.lastPick = editor.Renderer.Pick(int(x), int(y))
		if evt.Modifier&fyne.KeyModifierShift != 0 {
			editor.SelectedObjects.Add(g.lastPick.Selection...)
			editor.SetSelection(true, editor.SelectedObjects)
		} else if evt.Modifier&fyne.KeyModifierSuper != 0 {
			log.Printf("Subtracting game picking unimplemented")
		} else {
			editor.SelectObjects(true, g.lastPick.Selection...)
		}
	}

}
func (g *GameWidget) MouseUp(evt *desktop.MouseEvent) {}

func (g *GameWidget) requestFocus() {
	if c := fyne.CurrentApp().Driver().CanvasForObject(g); c != nil {
		c.Focus(g)
	}
}
