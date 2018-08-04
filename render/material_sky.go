package render

import (
	"image/color"
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/mapping/material"
	"github.com/tlyakhov/gofoom/registry"
)

type Sky material.Sky

func init() {
	registry.Instance().RegisterMapped(Sky{}, material.Sky{})
}

func (m *Sky) Sample(slice *Slice, u, v float64, light *concepts.Vector3, scale float64) color.NRGBA {
	v = float64(slice.Y) / (float64(slice.ScreenHeight) - 1)

	if m.StaticBackground {
		u = float64(slice.X) / (float64(slice.ScreenWidth) - 1)
	} else {
		u = float64(slice.Angle) / (2.0 * math.Pi)
	}
	return m.Sampler.Sample(u, v, 1.0)
}
