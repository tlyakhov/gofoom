package material

import (
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/materials"
	"github.com/tlyakhov/gofoom/render/state"
)

type SampledService struct {
	*materials.Sampled
	*state.Slice
}

func NewSampledService(m *materials.Sampled, s *state.Slice) *SampledService {
	return &SampledService{Sampled: m, Slice: s}
}

func (m *SampledService) Sample(u, v float64, light *concepts.Vector3, scale float64) uint32 {
	if m.IsLiquid {
		u += math.Cos(float64(m.Frame)*constants.LiquidChurnSpeed*concepts.Deg2rad) * constants.LiquidChurnSize
		v += math.Sin(float64(m.Frame)*constants.LiquidChurnSpeed*concepts.Deg2rad) * constants.LiquidChurnSize
	}

	for ; u < 0; u++ {
	}
	for ; u > 1; u-- {
	}
	for ; v < 0; v++ {
	}
	for ; v > 1; v-- {
	}

	return m.Sampler.Sample(u, v, scale)
}
