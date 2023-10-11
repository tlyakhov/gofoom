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

func (m *LitSampledService) Sample(u, v float64, light *concepts.Vector3, scale float64) uint32 {
	surface := concepts.Int32ToVector3(m.SampledService.Sample(u, v, light, scale))
	amb := *light
	amb.AddSelf(&m.Ambient)
	sum := (&surface).Mul3Self(&m.Diffuse).Mul3Self(&amb).ClampSelf(0.0, 255.0)
	return sum.ToInt32Color()
}
