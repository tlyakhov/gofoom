package main

import (
	"math"

	"tlyakhov/gofoom/editor/actions"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func MapMotionNotify(da *gtk.DrawingArea, ev *gdk.Event) {
	motion := gdk.EventMotionNewFromEvent(ev)
	x, y := motion.MotionVal()
	if x == editor.Mouse[0] && y == editor.Mouse[1] {
		return
	}
	editor.Mouse[0] = x
	editor.Mouse[1] = y
	editor.MouseWorld = *editor.ScreenToWorld(&editor.Mouse)
	editor.UpdateStatus()

	if editor.CurrentAction != nil {
		editor.CurrentAction.OnMouseMove()
	}
}

func MapButtonPress(da *gtk.DrawingArea, ev *gdk.Event) {
	press := gdk.EventButtonNewFromEvent(ev)
	editor.MousePressed = true
	editor.MouseDown[0], editor.MouseDown[1] = press.MotionVal()
	editor.MouseDownWorld = *editor.ScreenToWorld(&editor.MouseDown)

	da.GrabFocus()

	if press.Button() == 3 && editor.CurrentAction == nil {
		editor.NewAction(&actions.Select{IEditor: editor})
	} else if press.Button() == 2 && editor.CurrentAction == nil {
		editor.NewAction(&actions.Pan{IEditor: editor})
	} else if press.Button() == 1 && editor.CurrentAction == nil && len(editor.SelectedObjects) > 0 {
		editor.NewAction(&actions.Move{IEditor: editor})
	}

	if editor.CurrentAction != nil {
		editor.CurrentAction.OnMouseDown(press)
	}
}

func MapButtonRelease(da *gtk.DrawingArea, ev *gdk.Event) {
	//release := &gdk.EventButton{ev}
	editor.MousePressed = false

	if editor.CurrentAction != nil {
		editor.CurrentAction.OnMouseUp()
	}
}

func MapScroll(da *gtk.DrawingArea, ev *gdk.Event) {
	scroll := gdk.EventScrollNewFromEvent(ev)
	delta := 0.25
	if scroll.DeltaY() != 0 {
		delta = math.Abs(scroll.DeltaY() / 5)
	}

	if scroll.Direction() == gdk.SCROLL_DOWN {
		delta = -delta
	}
	if editor.Scale > 0.25 {
		editor.Scale += delta * 0.2
	} else if editor.Scale > 0.025 {
		editor.Scale += delta * 0.02
	} else if editor.Scale > 0.0025 {
		editor.Scale += delta * 0.002
	}
}

func GameButtonPress(da *gtk.DrawingArea, ev *gdk.Event) {
	press := gdk.EventButtonNewFromEvent(ev)
	da.GrabFocus()

	// TODO: make this more granular, and also support mobs
	if press.Button() == 1 {
		daw := da.GetAllocatedWidth()
		dah := da.GetAllocatedHeight()
		rw := editor.Renderer.ScreenWidth
		rh := editor.Renderer.ScreenHeight
		x := press.X() * float64(rw) / float64(daw)
		y := press.Y() * float64(rh) / float64(dah)
		picked := editor.Renderer.Pick(int(x), int(y))
		objects := make([]any, 0)
		for _, p := range picked {
			objects = append(objects, p.Attachable)
		}
		editor.SelectObjects(objects)
	}
}
