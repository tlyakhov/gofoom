package render

import (
	"image/color"

	"github.com/tlyakhov/gofoom/material"
	"github.com/tlyakhov/gofoom/math"
)

func (m *material.Lit) Sample(slice *Slice, u, v float64, light *math.Vector3, scaledHeight uint) color.NRGBA {
	surface := m.Sampled.Sample(slice, u, v, light, scaledHeight)
	sum := &math.Vector3{float64(surface.R), float64(surface.G), float64(surface.B)}
	sum = sum.Mul3(m.Diffuse)

	if light != nil {
		sum = sum.Mul3(light)
	}
	sum = sum.Add(m.Ambient).Clamp(0.0, 255.0)
	return color.NRGBA{uint8(sum.X), uint8(sum.Y), uint8(sum.Z), 255}
}
