// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"image"
	"math"
	"tlyakhov/gofoom/components/behaviors"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/containers"
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

func (mw *MapWidget) DrawQuadNode(node *core.QuadNode, index int) {
	switch index {
	case 0:
		mw.Context.SetRGBA(1.0, 0.0, 0.0, 0.25)
	case 1:
		mw.Context.SetRGBA(1.0, 0.6, 0.0, 0.25)
	case 2:
		mw.Context.SetRGBA(1.0, 0.0, 0.6, 0.25)
	case 3:
		mw.Context.SetRGBA(1.0, 0.6, 0.6, 0.25)
	}
	mw.Context.NewSubPath()
	mw.Context.MoveTo(node.Min[0], node.Min[1])
	mw.Context.LineTo(node.Max[0], node.Min[1])
	mw.Context.LineTo(node.Max[0], node.Max[1])
	mw.Context.LineTo(node.Min[0], node.Max[1])
	mw.Context.ClosePath()
	mw.Context.Stroke()
	text := fmt.Sprintf("r: %.2f, b: %v, l: %v", node.MaxRadius, len(node.Bodies), len(node.Lights))
	mw.Context.DrawStringAnchored(text, (node.Min[0]+node.Max[0])*0.5, (node.Min[1]+node.Max[1])*0.5, 0.5, 0.5)

	if !node.IsLeaf() {
		for i, child := range node.Children {
			mw.DrawQuadNode(child, i)
		}
	}
}

func (mw *MapWidget) render() {
	editor.GatherHoveringObjects()

	mw.Context.Identity()
	editor.MapViewGrid.Draw(&editor.EditorState)
	copy(mw.Context.Image().(*image.RGBA).Pix, editor.MapViewGrid.pixels())
	TransformContext(mw.Context)
	mw.Context.FontHeight()

	highlightedSectors := make(containers.Set[*core.Sector])
	for _, s := range editor.SelectedObjects.Exact {
		if s.Sector != nil && (s.Sector.Flags&ecs.ComponentHideInEditor) == 0 {
			highlightedSectors.Add(s.Sector)
		}
	}
	for _, s := range editor.HoveringObjects.Exact {
		if s.Sector != nil && (s.Sector.Flags&ecs.ComponentHideInEditor) == 0 {
			highlightedSectors.Add(s.Sector)
		}
	}

	colSector := ecs.ArenaFor[core.Sector](core.SectorCID)
	for i := range colSector.Cap() {
		if sector := colSector.Value(i); sector != nil && (sector.Flags&ecs.ComponentHideInEditor) == 0 {
			if _, ok := highlightedSectors[sector]; ok {
				continue
			}
			mw.DrawSector(sector)
		}
	}
	for sector := range highlightedSectors {
		mw.DrawSector(sector)
	}

	colSeg := ecs.ArenaFor[core.InternalSegment](core.InternalSegmentCID)
	for i := range colSeg.Cap() {
		if seg := colSeg.Value(i); seg != nil && (seg.Flags&ecs.ComponentHideInEditor) == 0 {
			mw.DrawInternalSegment(seg)
		}
	}
	colWaypoint := ecs.ArenaFor[behaviors.ActionWaypoint](behaviors.ActionWaypointCID)
	for i := range colWaypoint.Cap() {
		if waypoint := colWaypoint.Value(i); waypoint != nil && (waypoint.Flags&ecs.ComponentHideInEditor) == 0 {
			mw.DrawActions(waypoint.Entity)
		}
	}
	if editor.BodiesVisible {
		col3 := ecs.ArenaFor[core.Body](core.BodyCID)
		for i := range col3.Cap() {
			if body := col3.Value(i); body != nil && (body.Flags&ecs.ComponentHideInEditor) == 0 {
				mw.DrawBody(body)
			}
		}
	}

	// Quadtree testing code
	mw.DrawQuadNode(core.QuadTree.Root, 0)

	/*// Portal testing code
	p := core.GetBody(editor.Renderer.Player.Entity)
	v := &concepts.Vector2{p.Pos.Render[0], p.Pos.Render[1]}
	v2 := v.Add(&concepts.Vector2{math.Cos(p.Angle.Render*concepts.Deg2rad) * 10, math.Sin(p.Angle.Render*concepts.Deg2rad) * 10})
	mw.Context.SetRGBA(1.0, 0.0, 0.0, 1.0)
	mw.Context.NewSubPath()
	mw.Context.MoveTo(v[0], v[1])
	mw.Context.LineTo(v2[0], v2[1])
	mw.Context.ClosePath()
	mw.Context.Stroke()

	portalSector1 := core.GetSector(44)
	portalSegment1 := portalSector1.Segments[4]
	portalSector2 := core.GetSector(19)
	portalSegment2 := portalSector2.Segments[1]
	v3 := portalSegment1.PortalMatrix.Unproject(v)
	v3 = portalSegment2.MirrorPortalMatrix.Project(v3)
	a := p.Angle.Render - math.Atan2(portalSegment1.Normal[1], portalSegment1.Normal[0])*concepts.Rad2deg + math.Atan2(portalSegment2.Normal[1], portalSegment2.Normal[0])*concepts.Rad2deg + 180
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
}

func (mw *MapWidget) Draw(w, h int) image.Image {
	w /= state.MapViewRenderScale
	h /= state.MapViewRenderScale
	if mw.Context == nil || mw.Surface.Rect.Max.X != w || mw.Surface.Rect.Max.Y != h {
		editor.Lock.Lock()
		mw.Surface = image.NewRGBA(image.Rect(0, 0, w, h))
		mw.Context = gg.NewContext(w, h)
		editor.MapViewGrid.GridContext = gg.NewContext(w, h)
		editor.Size = concepts.Vector2{float64(w), float64(h)}
		editor.Lock.Unlock()
	}

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
		if ks, ok := action.Shortcut.(*desktop.CustomShortcut); ok {
			if action.NoModifier && evt.Name == ks.KeyName {
				action.Menu.Action()
			}
		}
	}
}

func (mw *MapWidget) TypedShortcut(s fyne.Shortcut) {
	switch s.ShortcutName() {
	case "Undo":
		editor.UndoCurrent()
	case "Redo":
		editor.RedoCurrent()
	case "Cut":
		editor.Act(&actions.Copy{Action: state.Action{IEditor: editor}, Cut: true})
	case "Copy":
		editor.Act(&actions.Copy{Action: state.Action{IEditor: editor}, Cut: false})
	case "Paste":
		editor.Act(&actions.Paste{Transform: actions.Transform{Action: state.Action{IEditor: editor}}})
	}
}

func (mw *MapWidget) setMouseDown(evt *fyne.PointEvent) bool {
	if mw.Context == nil {
		return false
	}
	mw.requestFocus()
	editor.MousePressed = true
	scale := float64(mw.Context.Width()) / float64(mw.Size().Width)
	editor.MouseDown[0], editor.MouseDown[1] = float64(evt.Position.X)*scale, float64(evt.Position.Y)*scale
	editor.MouseDownWorld = *editor.ScreenToWorld(&editor.MouseDown)
	return true
}

func (mw *MapWidget) setMouse(evt *fyne.PointEvent) bool {
	if mw.Context == nil {
		return false
	}
	scale := float64(mw.Context.Width()) / float64(mw.Size().Width)
	x, y := float64(evt.Position.X)*scale, float64(evt.Position.Y)*scale
	if x == editor.Mouse[0] && y == editor.Mouse[1] {
		return false
	}
	editor.Mouse[0], editor.Mouse[1] = x, y
	editor.MouseWorld = *editor.ScreenToWorld(&editor.Mouse)
	return true
}

func (mw *MapWidget) MouseDown(evt *desktop.MouseEvent) {
	//log.Printf("Mouse down")
	if !mw.setMouseDown(&evt.PointEvent) {
		return
	}

	switch {
	case evt.Button == desktop.MouseButtonSecondary && editor.CurrentAction == nil:
		editor.Act(&actions.Select{Place: actions.Place{Action: state.Action{IEditor: editor}}})
	case evt.Button == desktop.MouseButtonTertiary && editor.CurrentAction == nil:
		editor.Act(&actions.Pan{Action: state.Action{IEditor: editor}})
	case evt.Button == desktop.MouseButtonPrimary && editor.CurrentAction == nil && !editor.SelectedObjects.Empty():
		editor.Act(&actions.Transform{Action: state.Action{IEditor: editor}})
	}

	if placeable, ok := editor.CurrentAction.(actions.Placeable); ok {
		placeable.BeginPoint(evt.Modifier, evt.Button)
	}
	if m, ok := editor.CurrentAction.(desktop.Mouseable); ok {
		m.MouseDown(evt)
	}
}
func (mw *MapWidget) MouseUp(evt *desktop.MouseEvent) {
	editor.MousePressed = false

	//log.Printf("MouseUp")

	if m, ok := editor.CurrentAction.(desktop.Mouseable); ok {
		m.MouseUp(evt)
	}
	if bp, ok := editor.CurrentAction.(actions.Placeable); ok {
		bp.EndPoint()
	}
}

func (mw *MapWidget) Tapped(evt *fyne.PointEvent) {
	if !mw.setMouseDown(evt) {
		return
	}

	//log.Printf("Tapped")

	if m, ok := editor.CurrentAction.(fyne.Tappable); ok {
		m.Tapped(evt)
	}
	_, isDesktop := editor.App.(desktop.App)
	if bp, ok := editor.CurrentAction.(actions.Placeable); ok && !isDesktop {
		bp.BeginPoint(0, desktop.MouseButtonPrimary)
		bp.EndPoint()
	}
}

func (mw *MapWidget) TappedSecondary(evt *fyne.PointEvent) {
	if !mw.setMouseDown(evt) {
		return
	}

	//log.Printf("SecondaryTapped")

	if m, ok := editor.CurrentAction.(fyne.SecondaryTappable); ok {
		m.TappedSecondary(evt)
	}
	_, isDesktop := editor.App.(desktop.App)
	if bp, ok := editor.CurrentAction.(actions.Placeable); ok && !isDesktop {
		bp.BeginPoint(0, desktop.MouseButtonSecondary)
		bp.EndPoint()
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
	editor.MousePressed = false
	//log.Printf("DragEnd")

	if m, ok := editor.CurrentAction.(fyne.Draggable); ok {
		m.DragEnd()
	}
	if bp, ok := editor.CurrentAction.(actions.Placeable); ok {
		bp.EndPoint()
	}
}

func (mw *MapWidget) Dragged(evt *fyne.DragEvent) {
	if !mw.setMouse(&evt.PointEvent) {
		return
	}

	editor.Dragging = true

	//log.Printf("Dragged: %v", evt.Position)

	if bp, ok := editor.CurrentAction.(actions.Placeable); ok {
		if !bp.Point() {
			bp.BeginPoint(0, desktop.MouseButtonPrimary)
			bp.Point()
		}
	}
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
	if !mw.setMouse(&ev.PointEvent) {
		return
	}
	//log.Printf("scale:%v, x,y: %v, %v - world: %v, %v", scale, x, y, editor.MouseWorld[0], editor.MouseWorld[1])
	editor.UpdateStatus()
	//log.Printf("Moved")

	if bp, ok := editor.CurrentAction.(actions.Placeable); ok {
		bp.Point()
	}
	if m, ok := editor.CurrentAction.(desktop.Hoverable); ok {
		m.MouseMoved(ev)
	}
}

func (mw *MapWidget) Cursor() desktop.Cursor {
	return mw.MapCursor
}
