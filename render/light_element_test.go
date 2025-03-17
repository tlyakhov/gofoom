// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/ecs"
	"tlyakhov/gofoom/render"
)

const bounds = 30000

func BenchmarkLightmapConversion(b *testing.B) {
	b.Run("Correctness", func(b *testing.B) {
		s := new(core.Sector)
		s.Construct(nil)
		ecs := ecs.NewUniverse()
		c := render.Config{Universe: ecs}
		c.Initialize()
		v := new(concepts.Vector3)
		n := new(concepts.Vector3)
		result := new(concepts.Vector3)
		for i := 0; i < b.N; i++ {
			s.Min[0] = rand.Float64()*bounds - bounds/2
			s.Min[1] = rand.Float64()*bounds - bounds/2
			s.Min[2] = rand.Float64()*bounds - bounds/2
			v[0] = rand.Float64()*bounds + s.Min[0] - c.LightGrid
			v[1] = rand.Float64()*bounds + s.Min[1] - c.LightGrid
			v[2] = rand.Float64()*bounds + s.Min[2] - c.LightGrid
			n[0] = rand.Float64()*2.0 - 1.0
			n[1] = rand.Float64()*2.0 - 1.0
			n[2] = rand.Float64()*2.0 - 1.0
			n.NormSelf()

			a := c.WorldToLightmapHash(s, v, n)
			c.LightmapHashToWorld(s, result, a)

			dx := math.Floor(v[0]/constants.LightGrid)*c.LightGrid - result[0]
			dy := math.Floor(v[1]/constants.LightGrid)*c.LightGrid - result[1]
			dz := math.Floor(v[2]/constants.LightGrid)*c.LightGrid - result[2]
			if dx != 0 || dy != 0 || dz != 0 {
				fmt.Printf("Error: lightmap address conversion resulted in delta: %v,%v,%v\n", dx, dy, dz)
				break
			}
		}
	})
}
