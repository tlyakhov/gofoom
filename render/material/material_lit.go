package material

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/mapping/material"
	"github.com/tlyakhov/gofoom/render/state"
)

type LitService struct {
	*material.Lit
	*state.Slice
}

func NewLitService(m *material.Lit, s *state.Slice) *LitService {
	return &LitService{Lit: m, Slice: s}
}

func (m *LitService) Sample(u, v float64, light *concepts.Vector3, scale float64) uint32 {
	sum := m.Diffuse
	if light != nil {
		sum = sum.Mul3(light)
	}
	sum = sum.Add(m.Ambient).Clamp(0.0, 255.0)
	return sum.ToInt32Color()
}
