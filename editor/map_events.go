package main

import (
	"math"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func MapMotionNotify(da *gtk.DrawingArea, ev *gdk.Event) {
	motion := gdk.EventMotionNewFromEvent(ev)
	x, y := motion.MotionVal()
	if x == editor.Mouse.X && y == editor.Mouse.Y {
		return
	}
	editor.Mouse.X = x
	editor.Mouse.Y = y
	editor.MouseWorld = editor.ScreenToWorld(editor.Mouse)

	if editor.CurrentAction != nil {
		editor.CurrentAction.OnMouseMove()
	}
}

func MapButtonPress(da *gtk.DrawingArea, ev *gdk.Event) {
	press := gdk.EventButtonNewFromEvent(ev)
	editor.MousePressed = true
	editor.MouseDown.X, editor.MouseDown.Y = press.MotionVal()
	editor.MouseDownWorld = editor.ScreenToWorld(editor.MouseDown)

	if press.Button() == 3 && editor.CurrentAction == nil {
		editor.NewAction(&SelectAction{Editor: editor})
	} else if press.Button() == 2 && editor.CurrentAction == nil {
		editor.NewAction(&PanAction{Editor: editor})
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
	delta := math.Abs(scroll.DeltaY() / 5)
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
