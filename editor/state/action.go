// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

import "fyne.io/fyne/v2/driver/desktop"

// Actionable represents a generic editor action.
type Actionable interface {
	Act()
	Undo()
	Redo()
}

type MouseActionable interface {
	OnMouseDown(evt *desktop.MouseEvent)
	OnMouseUp()
	OnMouseMove()
	Cancel()
	Status() string
}
