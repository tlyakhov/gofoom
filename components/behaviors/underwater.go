// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/ecs"
)

type Underwater struct {
	ecs.Attached `editable:"^"`
}

func (u *Underwater) Shareable() bool { return true }
