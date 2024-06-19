// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"tlyakhov/gofoom/concepts"
)

type Underwater struct {
	concepts.Attached `editable:"^"`
}

var UnderwaterComponentIndex int

func init() {
	UnderwaterComponentIndex = concepts.DbTypes().Register(Underwater{}, UnderwaterFromDb)
}

func UnderwaterFromDb(entity *concepts.EntityRef) *Underwater {
	if asserted, ok := entity.Component(UnderwaterComponentIndex).(*Underwater); ok {
		return asserted
	}
	return nil
}
