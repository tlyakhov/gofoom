package material

import (
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

func (m *LitSampledService) Sample(u, v float64, light *concepts.Vector3, scale float64) uint32 {
	surface := concepts.Int32ToVector3(m.SampledService.Sample(u, v, light, scale))
	sum := &surface
	sum = sum.Mul3(m.Diffuse)

	if light != nil {
		sum = sum.Mul3(light)
	}
	sum = sum.Add(m.Ambient).Clamp(0.0, 255.0)
	return sum.ToInt32Color()
}
