package main

import (
	"image"
	"math"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/render"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/fogleman/gg"
)

// Declare conformity with interfaces
var _ fyne.Focusable = (*MapWidget)(nil)
var _ fyne.Widget = (*MapWidget)(nil)
var _ desktop.Mouseable = (*MapWidget)(nil)
var _ desktop.Keyable = (*MapWidget)(nil)

type MapWidget struct {
	widget.BaseWidget

	Raster       *canvas.Raster
	Context      *gg.Context
	Surface      *image.RGBA
	Font         *render.Font
	MapCursor    desktop.Cursor
	NeedsRefresh bool
}

func NewMapWidget() *MapWidget {
	mw := &MapWidget{MapCursor: desktop.DefaultCursor}
	mw.ExtendBaseWidget(mw)
	mw.Font, _ = render.NewFont("/Library/Fonts/Courier New.ttf", 24)
	mw.Raster = canvas.NewRaster(mw.Draw)
	return mw
}

func (mw *MapWidget) TransformContext() {
	t := editor.Pos.Mul(-editor.Scale).Add(editor.Size.Mul(0.5))
	mw.Context.Translate(t[0], t[1])
	mw.Context.Scale(editor.Scale, editor.Scale)
}

func (mw *MapWidget) Draw(w, h int) image.Image {
	if mw.Context == nil || mw.Surface.Rect.Max.X != w || mw.Surface.Rect.Max.Y != h {
		mw.Surface = image.NewRGBA(image.Rect(0, 0, w, h))
		mw.Context = gg.NewContext(w, h)
	}
	editor.Size = concepts.Vector2{float64(w), float64(h)}

	mw.Context.Identity()
	editor.MapViewGrid.Draw(&editor.Edit, mw)
	mw.TransformContext()
	mw.Context.FontHeight()

	for _, isector := range editor.DB.All(core.SectorComponentIndex) {
		mw.DrawSector(isector.(*core.Sector))
	}

	switch editor.CurrentAction.(type) {
	case *actions.Select:
		if editor.MousePressed {
			v1, v2 := editor.SelectionBox()
			mw.Context.DrawRectangle(v1[0], v1[1], v2[0]-v1[0], v2[1]-v1[1])
			mw.Context.SetRGBA(0.2, 0.2, 1.0, 0.3)
			mw.Context.Fill()
			mw.Context.SetRGBA(0.67, 0.67, 1.0, 0.3)
			mw.Context.Stroke()
		}
	case *actions.AddSector:
		gridMouse := editor.WorldGrid(&editor.MouseWorld)
		mw.Context.SetRGB(ColorSelectionPrimary[0], ColorSelectionPrimary[1], ColorSelectionPrimary[2])
		mw.DrawHandle(gridMouse)
	case *actions.SplitSector, *actions.SplitSegment, *actions.AlignGrid:
		gridMouse := editor.WorldGrid(&editor.MouseWorld)
		gridMouseDown := editor.WorldGrid(&editor.MouseDownWorld)
		mw.Context.SetRGB(ColorSelectionPrimary[0], ColorSelectionPrimary[1], ColorSelectionPrimary[2])
		mw.DrawHandle(gridMouse)
		if editor.MousePressed {
			mw.Context.NewSubPath()
			mw.Context.MoveTo(gridMouseDown[0], gridMouseDown[1])
			mw.Context.LineTo(gridMouse[0], gridMouse[1])
			mw.Context.ClosePath()
			mw.Context.Stroke()
			mw.DrawHandle(gridMouseDown)
		}

	}

	//cr.ShowText(fmt.Sprintf("%v, %v", Mouse[0], Mouse[1]))*/

	pixels := mw.Context.Image().(*image.RGBA).Pix
	copy(mw.Surface.Pix, pixels)

	return mw.Surface
}

func (mw *MapWidget) MinSize() fyne.Size {
	mw.ExtendBaseWidget(mw)
	return mw.BaseWidget.MinSize()
}

func (mw *MapWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(mw.Raster)
}

func (mw *MapWidget) KeyDown(evt *fyne.KeyEvent) {
}
func (mw *MapWidget) KeyUp(evt *fyne.KeyEvent) {
}

func (mw *MapWidget) FocusLost()       {}
func (mw *MapWidget) FocusGained()     {}
func (mw *MapWidget) TypedRune(r rune) {}
func (mw *MapWidget) TypedKey(evt *fyne.KeyEvent) {
}
func (mw *MapWidget) MouseDown(evt *desktop.MouseEvent) {
	mw.requestFocus()
	editor.MousePressed = true
	editor.MouseDown[0], editor.MouseDown[1] = float64(evt.Position.X), float64(evt.Position.Y)
	editor.MouseDownWorld = *editor.ScreenToWorld(&editor.MouseDown)

	if evt.Button == desktop.MouseButtonSecondary && editor.CurrentAction == nil {
		editor.NewAction(&actions.Select{IEditor: editor})
	} else if evt.Button == desktop.MouseButtonTertiary && editor.CurrentAction == nil {
		editor.NewAction(&actions.Pan{IEditor: editor})
	} else if evt.Button == desktop.MouseButtonPrimary && editor.CurrentAction == nil && len(editor.SelectedObjects) > 0 {
		editor.NewAction(&actions.Move{IEditor: editor})
	}

	if editor.CurrentAction != nil {
		editor.CurrentAction.OnMouseDown(evt)
	}
	mw.NeedsRefresh = true
}
func (mw *MapWidget) MouseUp(evt *desktop.MouseEvent) {
	editor.MousePressed = false

	if editor.CurrentAction != nil {
		editor.CurrentAction.OnMouseUp()
	}
	mw.NeedsRefresh = true
}

func (mw *MapWidget) requestFocus() {
	if c := fyne.CurrentApp().Driver().CanvasForObject(mw); c != nil {
		c.Focus(mw)
	}
}

func (mw *MapWidget) Scrolled(ev *fyne.ScrollEvent) {
	delta := 0.0
	if ev.Scrolled.DY != 0 {
		delta = float64(ev.Scrolled.DY)
		delta = math.Sqrt(math.Abs(delta))
		delta *= 0.1
		if ev.Scrolled.DY < 0 {
			delta = -delta
		}
	}

	if editor.Scale > 0.25 {
		editor.Scale += delta * 0.2
	} else if editor.Scale > 0.025 {
		editor.Scale += delta * 0.02
	} else if editor.Scale > 0.0025 {
		editor.Scale += delta * 0.002
	} else {
		editor.Scale += delta * 0.0002
	}

	mw.NeedsRefresh = true
}

func (mw *MapWidget) MouseIn(ev *desktop.MouseEvent) {
}
func (mw *MapWidget) MouseOut() {
}

func (mw *MapWidget) MouseMoved(ev *desktop.MouseEvent) {
	x, y := float64(ev.Position.X), float64(ev.Position.Y)
	if x == editor.Mouse[0] && y == editor.Mouse[1] {
		return
	}
	editor.Mouse[0] = x
	editor.Mouse[1] = y
	editor.MouseWorld = *editor.ScreenToWorld(&editor.Mouse)
	editor.UpdateStatus()

	if editor.CurrentAction != nil {
		editor.CurrentAction.OnMouseMove()
		mw.NeedsRefresh = true
	}
}

func (mw *MapWidget) Cursor() desktop.Cursor {
	return mw.MapCursor
}
