// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"math/rand"

	"testing"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/ecs"
	_ "tlyakhov/gofoom/scripting_symbols"
)

func setup() *ecs.ECS {
	db := ecs.NewECS()
	CreateTestWorld2(db)
	return db
}

func BenchmarkScriptedCode(b *testing.B) {
	db := setup()
	s := core.Script{ECS: db}
	s.Construct(map[string]any{
		"Code":  "core.GetSector(sectorEntity).Bottom.Z.Original=5",
		"Style": "ScriptStyleStatement",
	})
	s.Vars["sector"] = db.GetEntityByName("sector1")
	b.Run("Script", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s.Act()
		}
	})
	b.Run("Native", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			core.GetSector(db, s.Vars["sector"].(ecs.Entity)).Bottom.Z.Original = rand.Float64()
		}
	})
}
