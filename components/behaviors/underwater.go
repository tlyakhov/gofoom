// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/ecs"
)

type Underwater struct {
	ecs.Attached `editable:"^"`
}

var UnderwaterComponentIndex int

func init() {
	UnderwaterComponentIndex = ecs.RegisterComponent(&ecs.ComponentColumn[Underwater, *Underwater]{Getter: GetUnderwater})
}

func GetUnderwater(db *ecs.ECS, e ecs.Entity) *Underwater {
	if asserted, ok := db.Component(e, UnderwaterComponentIndex).(*Underwater); ok {
		return asserted
	}
	return nil
}
