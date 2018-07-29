package render

import (
	"image/color"

	"github.com/tlyakhov/gofoom/material"
	"github.com/tlyakhov/gofoom/math"
)

func (m *material.Sky) Sample(slice *Slice, u, v float64, light *math.Vector3, scaledHeight uint) color.NRGBA {
	v = float64(slice.Y) / (float64(slice.ScreenHeight) - 1)

	if m.StaticBackground {
		u = float64(slice.X) / (float64(slice.ScreenWidth) - 1)
	} else {
		u = float64(slice.RayIndex) / (float64(slice.TrigCount) - 1)
	}
	// Assume largest mipmap
	scaledHeight = 0

	return m.Sampler.Sample(u, v, scaledHeight)
}
