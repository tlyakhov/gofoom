package material

import (
	"image/color"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/mapping/material"
	"github.com/tlyakhov/gofoom/render/state"
)

type LitSampledService struct {
	*material.LitSampled
	*SampledService
	*state.Slice
}

func NewLitSampledService(m *material.LitSampled, s *state.Slice) *LitSampledService {
	return &LitSampledService{
		LitSampled:     m,
		SampledService: NewSampledService(&m.Sampled, s),
		Slice:          s,
	}
}

func (m *LitSampledService) Sample(u, v float64, light *concepts.Vector3, scale float64) color.NRGBA {
	surface := m.SampledService.Sample(u, v, light, scale)
	sum := &concepts.Vector3{float64(surface.R), float64(surface.G), float64(surface.B)}
	sum = sum.Mul3(m.Diffuse)

	if light != nil {
		sum = sum.Mul3(light)
	}
	sum = sum.Add(m.Ambient).Clamp(0.0, 255.0)
	return color.NRGBA{uint8(sum.X), uint8(sum.Y), uint8(sum.Z), 255}
}
