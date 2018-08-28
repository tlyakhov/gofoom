package state

import (
	"github.com/gotk3/gotk3/gdk"
)

// IAction represents a generic editor action.
type IAction interface {
	OnMouseDown(button *gdk.EventButton)
	OnMouseUp()
	OnMouseMove()
	Act()
	Cancel()
	Frame()
	Undo()
	Redo()
}
