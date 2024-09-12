// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"image"
	"log"
	"math"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/actions"
	"tlyakhov/gofoom/editor/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/fogleman/gg"
)

// Declare conformity with interfaces
var _ fyne.Draggable = (*MapWidget)(nil)
var _ fyne.Focusable = (*MapWidget)(nil)
var _ fyne.Tappable = (*MapWidget)(nil)
var _ fyne.SecondaryTappable = (*MapWidget)(nil)
var _ fyne.Widget = (*MapWidget)(nil)
var _ desktop.Mouseable = (*MapWidget)(nil)
var _ desktop.Keyable = (*MapWidget)(nil)

type MapWidget struct {
	widget.BaseWidget

	Raster    *canvas.Raster
	Context   *gg.Context
	Surface   *image.RGBA
	MapCursor desktop.Cursor
}

func NewMapWidget() *MapWidget {
	mw := &MapWidget{MapCursor: desktop.DefaultCursor}
	mw.ExtendBaseWidget(mw)
	mw.Raster = canvas.NewRaster(mw.Draw)
	return mw
}

func TransformContext(context *gg.Context) {
	t := editor.Pos.Mul(-editor.Scale).Add(editor.Size.Mul(0.5))
	context.Translate(t[0], t[1])
	context.Scale(editor.Scale, editor.Scale)
}

func (mw *MapWidget) Draw(w, h int) image.Image {
	editor.Lock.Lock()
	defer editor.Lock.Unlock()

	w /= state.MapViewRenderScale
	h /= state.MapViewRenderScale
	if mw.Context == nil || mw.Surface.Rect.Max.X != w || mw.Surface.Rect.Max.Y != h {
		mw.Surface = image.NewRGBA(image.Rect(0, 0, w, h))
		mw.Context = gg.NewContext(w, h)
		editor.MapViewGrid.GridContext = gg.NewContext(w, h)
		editor.Size = concepts.Vector2{float64(w), float64(h)}
	}

	mw.Context.Identity()
	editor.MapViewGrid.Draw(&editor.Edit)
	copy(mw.Context.Image().(*image.RGBA).Pix, editor.MapViewGrid.pixels())
	TransformContext(mw.Context)
	mw.Context.FontHeight()

	// Highlight PVS sectors
	pvsSector := make(map[ecs.Entity]*core.Sector)
	for _, s := range editor.SelectedObjects.Exact {
		if s.Sector == nil {
			continue
		}
		for entity, sector := range s.Sector.PVS {
			pvsSector[entity] = sector
		}
	}

	colSector := ecs.ColumnFor[core.Sector](editor.ECS, core.SectorCID)
	for i := range colSector.Length {
		mw.DrawSector(colSector.Value(i), pvsSector[colSector.Value(i).Entity] != nil)
	}
	colSeg := ecs.ColumnFor[core.InternalSegment](editor.ECS, core.InternalSegmentCID)
	for i := range colSeg.Length {
		mw.DrawInternalSegment(colSeg.Value(i))
	}
	colWaypoint := ecs.ColumnFor[behaviors.ActionWaypoint](editor.ECS, behaviors.ActionWaypointCID)
	for i := range colWaypoint.Length {
		mw.DrawActions(colWaypoint.Value(i).Entity)
	}
	if editor.BodiesVisible {
		col3 := ecs.ColumnFor[core.Body](editor.ECS, core.BodyCID)
		for i := range col3.Length {
			mw.DrawBody(col3.Value(i))
		}
	}
	/*// Portal testing code
	p := core.GetBody(editor.ECS, editor.Renderer.Player.Entity)
	v := &concepts.Vector2{p.Pos.Now[0], p.Pos.Now[1]}
	v2 := v.Add(&concepts.Vector2{math.Cos(p.Angle.Now*concepts.Deg2rad) * 10, math.Sin(p.Angle.Now*concepts.Deg2rad) * 10})
	mw.Context.SetRGBA(1.0, 0.0, 0.0, 1.0)
	mw.Context.NewSubPath()
	mw.Context.MoveTo(v[0], v[1])
	mw.Context.LineTo(v2[0], v2[1])
	mw.Context.ClosePath()
	mw.Context.Stroke()

	portalSector1 := core.GetSector(editor.ECS, 44)
	portalSegment1 := portalSector1.Segments[4]
	portalSector2 := core.GetSector(editor.ECS, 19)
	portalSegment2 := portalSector2.Segments[1]
	v3 := portalSegment1.PortalMatrix.Unproject(v)
	v3 = portalSegment2.MirrorPortalMatrix.Project(v3)
	a := p.Angle.Now - math.Atan2(portalSegment1.Normal[1], portalSegment1.Normal[0])*concepts.Rad2deg + math.Atan2(portalSegment2.Normal[1], portalSegment2.Normal[0])*concepts.Rad2deg + 180
	v4 := v3.Add(&concepts.Vector2{math.Cos(a*concepts.Deg2rad) * 10, math.Sin(a*concepts.Deg2rad) * 10})

	mw.Context.SetRGBA(1.0, 0.0, 0.0, 1.0)
	mw.Context.NewSubPath()
	mw.Context.MoveTo(v3[0], v3[1])
	mw.Context.LineTo(v4[0], v4[1])
	mw.Context.ClosePath()
	mw.Context.Stroke()*/

	switch editor.CurrentAction.(type) {
	case *actions.Select:
		if editor.MousePressed {
			v1, v2 := editor.SelectionBox()
			mw.Context.DrawRectangle(v1[0], v1[1], v2[0]-v1[0], v2[1]-v1[1])
			mw.Context.SetRGBA(0.2, 0.2, 1.0, 0.3)
			mw.Context.FillPreserve()
			mw.Context.SetRGBA(0.67, 0.67, 1.0, 0.3)
			mw.Context.Stroke()
		}
	case *actions.Transform:
		gridMouse := editor.WorldGrid(&editor.MouseWorld)
		gridMouseDown := editor.WorldGrid(&editor.MouseDownWorld)
		mw.Context.SetRGBA(0.2, 0.6, 0.7, 0.8)
		mw.Context.DrawRectangle(gridMouseDown[0]-2, gridMouseDown[1]-2, 4, 4)
		mw.Context.Fill()
		mw.Context.DrawRectangle(gridMouse[0]-2, gridMouse[1]-2, 4, 4)
		mw.Context.Fill()
		t := editor.CurrentAction.(*actions.Transform)
		var label string
		switch t.Mode {
		case "rotate":
			label = fmt.Sprintf("Rotating: %.1f", t.Angle*180/math.Pi)
		case "rotate-constrained":
			factor := math.Pi * 0.25
			label = fmt.Sprintf("Rotating: %.1f", math.Round(t.Angle/factor)*factor*180/math.Pi)
		default:
			label = fmt.Sprintf("Translating: %.1f, %.1f", t.Delta[0], t.Delta[1])
		}
		mw.Context.DrawStringAnchored(label, gridMouseDown[0], gridMouseDown[1], 0.5, 1.0)
	case *actions.AddSector:
		gridMouse := editor.WorldGrid(&editor.MouseWorld)
		mw.Context.SetStrokeStyle(PatternSelectionPrimary)
		mw.DrawHandle(gridMouse)
	case *actions.SplitSector, *actions.SplitSegment, *actions.AlignGrid:
		gridMouse := editor.WorldGrid(&editor.MouseWorld)
		gridMouseDown := editor.WorldGrid(&editor.MouseDownWorld)
		mw.Context.SetStrokeStyle(PatternSelectionPrimary)
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
	editor.KeysDown.Add(evt.Name)
}
func (mw *MapWidget) KeyUp(evt *fyne.KeyEvent) {
	editor.KeysDown.Delete(evt.Name)
}

func (mw *MapWidget) FocusLost()       {}
func (mw *MapWidget) FocusGained()     {}
func (mw *MapWidget) TypedRune(r rune) {}
func (mw *MapWidget) TypedKey(evt *fyne.KeyEvent) {
	for _, action := range editor.MenuActions {
		if action.NoModifier && evt.Name == action.Shortcut.KeyName {
			action.Menu.Action()
		}
	}
}
func (mw *MapWidget) MouseDown(evt *desktop.MouseEvent) {
	if mw.Context == nil {
		return
	}
	log.Printf("Mouse down")
	mw.requestFocus()
	editor.MousePressed = true
	scale := float64(mw.Context.Width()) / float64(mw.Size().Width)
	editor.MouseDown[0], editor.MouseDown[1] = float64(evt.Position.X)*scale, float64(evt.Position.Y)*scale
	editor.MouseDownWorld = *editor.ScreenToWorld(&editor.MouseDown)

	if evt.Button == desktop.MouseButtonSecondary && editor.CurrentAction == nil {
		editor.NewAction(&actions.Select{IEditor: editor})
	} else if evt.Button == desktop.MouseButtonTertiary && editor.CurrentAction == nil {
		editor.NewAction(&actions.Pan{IEditor: editor})
	} else if evt.Button == desktop.MouseButtonPrimary && editor.CurrentAction == nil && !editor.SelectedObjects.Empty() {
		editor.NewAction(&actions.Transform{IEditor: editor})
	}

	if m, ok := editor.CurrentAction.(desktop.Mouseable); ok {
		m.MouseDown(evt)
	}
}
func (mw *MapWidget) MouseUp(evt *desktop.MouseEvent) {
	editor.MousePressed = false

	if m, ok := editor.CurrentAction.(desktop.Mouseable); ok {
		m.MouseUp(evt)
	}
}

func (mw *MapWidget) Tapped(evt *fyne.PointEvent) {
	if mw.Context == nil {
		return
	}
	mw.requestFocus()

	scale := float64(mw.Context.Width()) / float64(mw.Size().Width)
	editor.MouseDown[0], editor.MouseDown[1] = float64(evt.Position.X)*scale, float64(evt.Position.Y)*scale
	editor.MouseDownWorld = *editor.ScreenToWorld(&editor.MouseDown)
	log.Printf("Tapped")

	if m, ok := editor.CurrentAction.(fyne.Tappable); ok {
		m.Tapped(evt)
	}
}

func (mw *MapWidget) TappedSecondary(evt *fyne.PointEvent) {
	if mw.Context == nil {
		return
	}
	mw.requestFocus()

	scale := float64(mw.Context.Width()) / float64(mw.Size().Width)
	editor.MouseDown[0], editor.MouseDown[1] = float64(evt.Position.X)*scale, float64(evt.Position.Y)*scale
	editor.MouseDownWorld = *editor.ScreenToWorld(&editor.MouseDown)
	log.Printf("SecondaryTapped")

	if m, ok := editor.CurrentAction.(fyne.SecondaryTappable); ok {
		m.TappedSecondary(evt)
	}
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
		delta = math.Log(math.Abs(delta))
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
}

func (mw *MapWidget) DragEnd() {
	editor.Dragging = false
	//log.Printf("DragEnd")

	if m, ok := editor.CurrentAction.(fyne.Draggable); ok {
		m.DragEnd()
	}
}

func (mw *MapWidget) Dragged(evt *fyne.DragEvent) {
	if mw.Context == nil {
		return
	}

	editor.Dragging = true
	scale := float64(mw.Context.Width()) / float64(mw.Size().Width)
	editor.Mouse[0], editor.Mouse[1] = float64(evt.Position.X)*scale, float64(evt.Position.Y)*scale
	editor.MouseWorld = *editor.ScreenToWorld(&editor.Mouse)
	//log.Printf("Dragged: %v", evt.Position)

	if m, ok := editor.CurrentAction.(fyne.Draggable); ok {
		m.Dragged(evt)
	}
}

func (mw *MapWidget) MouseIn(ev *desktop.MouseEvent) {
	if m, ok := editor.CurrentAction.(desktop.Hoverable); ok {
		m.MouseIn(ev)
	}
}
func (mw *MapWidget) MouseOut() {
	if m, ok := editor.CurrentAction.(desktop.Hoverable); ok {
		m.MouseOut()
	}
}

func (mw *MapWidget) MouseMoved(ev *desktop.MouseEvent) {
	if mw.Context == nil {
		return
	}

	scale := float64(mw.Context.Width()) / float64(mw.Size().Width)
	x, y := float64(ev.Position.X)*scale, float64(ev.Position.Y)*scale
	if x == editor.Mouse[0] && y == editor.Mouse[1] {
		return
	}
	editor.Mouse[0], editor.Mouse[1] = x, y
	editor.MouseWorld = *editor.ScreenToWorld(&editor.Mouse)
	//log.Printf("scale:%v, x,y: %v, %v - world: %v, %v", scale, x, y, editor.MouseWorld[0], editor.MouseWorld[1])
	editor.UpdateStatus()
	//log.Printf("Moved")

	if m, ok := editor.CurrentAction.(desktop.Hoverable); ok {
		m.MouseMoved(ev)
	}
}

func (mw *MapWidget) Cursor() desktop.Cursor {
	return mw.MapCursor
}
