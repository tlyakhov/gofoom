// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package scripting_symbols

import (
	"math/rand"

	"testing"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/controllers"
	"tlyakhov/gofoom/ecs"
)

func setup() *ecs.Universe {
	u := ecs.NewUniverse()
	controllers.CreateTestWorld2(u)
	return u
}

func BenchmarkScriptedCode(b *testing.B) {
	u := setup()
	s := core.Script{Universe: u}
	s.Construct(map[string]any{
		"Code":  "core.GetSector(sectorEntity).Bottom.Z.Spawn=5",
		"Style": "ScriptStyleStatement",
	})
	s.Vars["sector"] = u.GetEntityByName("sector1")
	b.Run("Script", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s.Act()
		}
	})
	b.Run("Native", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			core.GetSector(u, s.Vars["sector"].(ecs.Entity)).Bottom.Z.Spawn = rand.Float64()
		}
	})
}
