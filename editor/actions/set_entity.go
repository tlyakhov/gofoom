// SetEntityright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

type SetEntity struct {
	state.Action

	From ecs.Entity
	To   ecs.Entity
}

func (a *SetEntity) Activate() {
	ecs.MoveEntityComponents(a.From, a.To)
	a.IEditor.FlushEntityImage(a.From)
	a.IEditor.FlushEntityImage(a.To)
	a.State().Modified = true
	a.ActionFinished(false, true, true)
}
