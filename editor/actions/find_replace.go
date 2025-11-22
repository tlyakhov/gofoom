// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"log"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/editor/state"
)

type FindReplace struct {
	state.Action

	From, To ecs.Entity
}

func (a *FindReplace) Activate() {
	defer a.ActionFinished(false, true, false)

	if a.From == a.To || a.From == 0 || a.To == 0 {
		log.Printf("Actions.FindReplace error.")
		return
	}
	ecs.FindReplaceRelations(a.From, a.To)
	a.IEditor.FlushEntityImage(a.From)
	a.IEditor.FlushEntityImage(a.To)
	a.State().Modified = true
}
