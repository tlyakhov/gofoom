// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"math/rand"
	"testing"
)

type mockC1 struct {
	Attached
}

type mockC2 struct {
	Attached
}

type mockC3 struct {
	Attached
}

type mockC4 struct {
	Attached
}

type mockC5 struct {
	Attached
}

type mockC6 struct {
	Attached
}

type mockC7 struct {
	Attached
}

type mockC8 struct {
	Attached
}

func BenchmarkGet(b *testing.B) {
	RegisterComponent(&Column[mockC1, *mockC1]{})
	RegisterComponent(&Column[mockC2, *mockC2]{})
	RegisterComponent(&Column[mockC3, *mockC3]{})
	RegisterComponent(&Column[mockC4, *mockC4]{})
	RegisterComponent(&Column[mockC5, *mockC5]{})
	RegisterComponent(&Column[mockC6, *mockC6]{})
	RegisterComponent(&Column[mockC7, *mockC7]{})
	RegisterComponent(&Column[mockC8, *mockC8]{})
	db := NewECS()
	cp := Types().ColumnPlaceholders
	numEntities := 1000
	for range numEntities {
		entity := db.NewEntity()
		for range 5 {
			index := rand.Intn(len(cp)-1) + 1
			db.NewAttachedComponent(entity, cp[index].ID())
		}
	}
	// Divide the result by 1000
	b.Run("Get", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			b.StopTimer()
			index := rand.Intn(len(cp)-1) + 1
			entity := Entity(rand.Intn(numEntities-1) + 1)
			cid := cp[index].ID()
			b.StartTimer()
			for range 1000 {
				db.Component(entity, cid)
			}
		}
	})
}
