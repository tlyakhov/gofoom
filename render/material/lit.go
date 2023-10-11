package material

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/materials"
	"tlyakhov/gofoom/render/state"
)

type LitService struct {
	*materials.Lit
	*state.Slice
}

func NewLitService(m *materials.Lit, s *state.Slice) *LitService {
	return &LitService{Lit: m, Slice: s}
}

func (m *LitService) Sample(u, v float64, light *concepts.Vector3, scale float64) uint32 {
	sum := m.Diffuse.Mul3(light.Add(&m.Ambient)).Mul(255.0).Clamp(0.0, 255.0)
	return sum.ToInt32Color()
}
