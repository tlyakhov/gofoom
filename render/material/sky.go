package material

import (
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/materials"
	"github.com/tlyakhov/gofoom/render/state"
)

type SkyService struct {
	*materials.Sky
	*state.Slice
}

func NewSkyService(m *materials.Sky, s *state.Slice) *SkyService {
	return &SkyService{Sky: m, Slice: s}
}

func (m *SkyService) Sample(u, v float64, light *concepts.Vector3, scale float64) uint32 {
	v = float64(m.Y) / (float64(m.ScreenHeight) - 1)

	if m.StaticBackground {
		u = float64(m.X) / (float64(m.ScreenWidth) - 1)
	} else {
		u = float64(m.Angle) / (2.0 * math.Pi)
	}
	return m.Sampler.Sample(u, v, 1.0)
}
