package state

import "fyne.io/fyne/v2/driver/desktop"

// IAction represents a generic editor action.
type IAction interface {
	OnMouseDown(evt *desktop.MouseEvent)
	OnMouseUp()
	OnMouseMove()
	Act()
	Cancel()
	Frame()
	Undo()
	Redo()
	RequiresLock() bool
}
