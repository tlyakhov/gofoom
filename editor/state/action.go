// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package state

// Actionable represents a generic editor action.
type Actionable interface {
	Act()
	Undo()
	Redo()
}

type Cancelable interface {
	Cancel()
	Status() string
}
