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

func (m *mockC1) ComponentID() ComponentID {
	return 1
}

type mockC2 struct {
	Attached
}

func (m *mockC2) ComponentID() ComponentID {
	return 2
}

type mockC3 struct {
	Attached
}

func (m *mockC3) ComponentID() ComponentID {
	return 3
}

type mockC4 struct {
	Attached
}

func (m *mockC4) ComponentID() ComponentID {
	return 4
}

type mockC5 struct {
	Attached
}

func (m *mockC5) ComponentID() ComponentID {
	return 5
}

type mockC6 struct {
	Attached
}

func (m *mockC6) ComponentID() ComponentID {
	return 6
}

type mockC7 struct {
	Attached
}

func (m *mockC7) ComponentID() ComponentID {
	return 7
}

type mockC8 struct {
	Attached
}

func (m *mockC8) ComponentID() ComponentID {
	return 8
}
func BenchmarkGet(b *testing.B) {
	RegisterComponent(&Arena[mockC1, *mockC1]{})
	RegisterComponent(&Arena[mockC2, *mockC2]{})
	RegisterComponent(&Arena[mockC3, *mockC3]{})
	RegisterComponent(&Arena[mockC4, *mockC4]{})
	RegisterComponent(&Arena[mockC5, *mockC5]{})
	RegisterComponent(&Arena[mockC6, *mockC6]{})
	RegisterComponent(&Arena[mockC7, *mockC7]{})
	RegisterComponent(&Arena[mockC8, *mockC8]{})
	cp := Types().ArenaPlaceholders
	numEntities := 1000
	for range numEntities {
		entity := NewEntity()
		for range 5 {
			index := rand.Intn(len(cp)-1) + 1
			NewAttachedComponent(entity, cp[index].ID())
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
				Component(entity, cid)
			}
		}
	})
}
