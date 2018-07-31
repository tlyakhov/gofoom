package render

import (
	"image/color"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/mapping/material"
)

type Sky material.Sky

func (m *Sky) Sample(slice *Slice, u, v float64, light *concepts.Vector3, scale float64) color.NRGBA {
	v = float64(slice.Y) / (float64(slice.ScreenHeight) - 1)

	if m.StaticBackground {
		u = float64(slice.X) / (float64(slice.ScreenWidth) - 1)
	} else {
		u = float64(slice.RayIndex) / (float64(slice.TrigCount) - 1)
	}
	return m.Sampler.Sample(u, v, 1.0)
}
