package render

import (
	"image/color"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/mapping/material"
	"github.com/tlyakhov/gofoom/registry"
)

type LitSampled material.LitSampled

func init() {
	registry.Instance().RegisterMapped(LitSampled{}, material.LitSampled{})
}

func (m *LitSampled) Sample(slice *Slice, u, v float64, light *concepts.Vector3, scale float64) color.NRGBA {
	sampled := (*Sampled)(m.Sampled)

	surface := sampled.Sample(slice, u, v, light, scale)
	sum := &concepts.Vector3{float64(surface.R), float64(surface.G), float64(surface.B)}
	sum = sum.Mul3(m.Diffuse)

	if light != nil {
		sum = sum.Mul3(light)
	}
	sum = sum.Add(m.Ambient).Clamp(0.0, 255.0)
	//fmt.Printf("sample: %v\n", m.Diffuse)
	//panic("asdf")
	return color.NRGBA{uint8(sum.X), uint8(sum.Y), uint8(sum.Z), 255}
}
