package material

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/materials"
	"github.com/tlyakhov/gofoom/render/state"
)

type LitService struct {
	*materials.Lit
	*state.Slice
}

func NewLitService(m *materials.Lit, s *state.Slice) *LitService {
	return &LitService{Lit: m, Slice: s}
}

func (m *LitService) Sample(u, v float64, light concepts.Vector3, scale float64) uint32 {
	sum := m.Diffuse.Mul3(light).Add(m.Ambient).Clamp(0.0, 255.0)
	return sum.ToInt32Color()
}
