package main

import (
	"github.com/gotk3/gotk3/gdk"
)

// AbstractAction represents a generic editor action.
type AbstractAction interface {
	OnMouseDown(button *gdk.EventButton)
	OnMouseUp()
	OnMouseMove()
	Act()
	Cancel()
	Frame()
	Undo()
	Redo()
}
