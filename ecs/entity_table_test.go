// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"math/rand/v2"
	"testing"
)

func BenchmarkEntityTable(b *testing.B) {
	table := make(EntityTable, 0)
	b.Run("Fuzz", func(b *testing.B) {
		for range b.N {
			table.Set(Entity(rand.Int() % 1000))
			table.Contains(Entity(rand.Int() % 1000))
			if rand.Int()%10 == 0 {
				table.Delete(Entity(rand.Int() % 1000))
			}
		}
	})
}
