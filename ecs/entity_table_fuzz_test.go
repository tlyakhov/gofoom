// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"math/rand/v2"
	"testing"
)

func TestEntityTableFuzz(t *testing.T) {
	table := make(EntityTable, 0)
	for range 1000 {
		table.Set(Entity(rand.Int() % 1000))
		table.Contains(Entity(rand.Int() % 1000))
		if rand.Int()%10 == 0 {
			table.Delete(Entity(rand.Int() % 1000))
		}
	}
}
