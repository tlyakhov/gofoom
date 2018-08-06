package material

import (
	"image/color"
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/mapping/material"
	"github.com/tlyakhov/gofoom/render/state"
)

type SkyService struct {
	*material.Sky
	*state.Slice
}

func NewSkyService(m *material.Sky, s *state.Slice) *SkyService {
	return &SkyService{Sky: m, Slice: s}
}

func (m *SkyService) Sample(u, v float64, light *concepts.Vector3, scale float64) color.NRGBA {
	v = float64(m.Y) / (float64(m.ScreenHeight) - 1)

	if m.StaticBackground {
		u = float64(m.X) / (float64(m.ScreenWidth) - 1)
	} else {
		u = float64(m.Angle) / (2.0 * math.Pi)
	}
	return m.Sampler.Sample(u, v, 1.0)
}
