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
	UnderwaterCID = ecs.RegisterComponent(&ecs.Arena[Underwater, *Underwater]{})
}

func (x *Underwater) ComponentID() ecs.ComponentID {
	return UnderwaterCID
}
func GetUnderwater(e ecs.Entity) *Underwater {
	if asserted, ok := ecs.Component(e, UnderwaterCID).(*Underwater); ok {
		return asserted
	}
	return nil
}

func (u *Underwater) MultiAttachable() bool { return true }
