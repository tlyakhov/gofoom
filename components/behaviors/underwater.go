// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/ecs"
)

type Underwater struct {
	ecs.Attached `editable:"^"`
}

var UnderwaterCID ecs.ComponentID

func init() {
	UnderwaterCID = ecs.RegisterComponent(&ecs.Column[Underwater, *Underwater]{Getter: GetUnderwater})
}

func GetUnderwater(db *ecs.ECS, e ecs.Entity) *Underwater {
	if asserted, ok := db.Component(e, UnderwaterCID).(*Underwater); ok {
		return asserted
	}
	return nil
}

func (u *Underwater) MultiAttachable() bool { return true }
