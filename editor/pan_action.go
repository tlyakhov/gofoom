package main

import (
	"fmt"

	"github.com/gotk3/gotk3/gdk"
	"github.com/tlyakhov/gofoom/concepts"
)

type PanAction struct {
	*Editor
	OriginalPos concepts.Vector2
	Delta       concepts.Vector2
}

func (a *PanAction) OnMouseDown(button *gdk.EventButton) {
	if button.State()&uint(gdk.GDK_SHIFT_MASK) != 0 {
		fmt.Println("asdfasdf")
	}
	a.State = "PanStart"
	a.OriginalPos = a.Pos
	// Set Cursor all-scroll
}

func (a *PanAction) OnMouseMove() {
	a.State = "Panning"
	a.Delta = a.Mouse.Sub(a.MouseDown)
	a.Pos = a.OriginalPos.Sub(a.Delta)
}

func (a *PanAction) OnMouseUp() {
	a.ActionFinished()
}
func (a *PanAction) Act()    {}
func (a *PanAction) Cancel() {}
func (a *PanAction) Frame()  {}

func (a *PanAction) Undo() {
	a.Pos = a.OriginalPos
}
func (a *PanAction) Redo() {
	a.Pos = a.OriginalPos.Sub(a.Delta)
}
