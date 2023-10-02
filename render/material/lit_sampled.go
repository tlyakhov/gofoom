package material

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/materials"
	"tlyakhov/gofoom/render/state"
)

type LitSampledService struct {
	*materials.LitSampled
	*SampledService
	*state.Slice
}

func NewLitSampledService(m *materials.LitSampled, s *state.Slice) *LitSampledService {
	return &LitSampledService{
		LitSampled:     m,
		SampledService: NewSampledService(&m.Sampled, s),
		Slice:          s,
	}
}

func (m *LitSampledService) Sample(u, v float64, light concepts.Vector3, scale float64) uint32 {
	surface := concepts.Int32ToVector3(m.SampledService.Sample(u, v, light, scale))
	sum := surface
	sum = sum.Mul3(m.Diffuse).Mul3(light.Add(m.Ambient)).Clamp(0.0, 255.0)
	return sum.ToInt32Color()
}
