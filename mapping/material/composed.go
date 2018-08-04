package material

import (
	"fmt"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/registry"
)

type LitSampled struct {
	*concepts.Base
	*Lit
	*Sampled
}

func init() {
	registry.Instance().Register(LitSampled{})
	registry.Instance().Register(PainfulLitSampled{})
}

func (m *LitSampled) Initialize() {
	m.Sampled = &Sampled{}
	m.Sampled.Initialize()
	m.Lit = &Lit{}
	m.Lit.Initialize()
	m.Lit.Base = m.Sampled.Base
	m.Base = m.Sampled.Base
}

func (m *LitSampled) Deserialize(data map[string]interface{}) {
	m.Initialize()
	m.Lit.Deserialize(data)
	m.Sampled.Deserialize(data)
	m.Lit.Base = m.Sampled.Base
	m.Base = m.Sampled.Base
}

type PainfulLitSampled struct {
	*concepts.Base
	*LitSampled
	*Painful
}

func (m *PainfulLitSampled) Initialize() {
	m.LitSampled = &LitSampled{}
	m.LitSampled.Initialize()
	m.Base = m.LitSampled.Base
	m.Painful = &Painful{}
}

func (m *PainfulLitSampled) Deserialize(data map[string]interface{}) {
	m.Initialize()
	m.LitSampled.Deserialize(data)
	m.Base = m.LitSampled.Base
	fmt.Printf("PainfulLitSampled: %v\n", m.Lit.ID)
}
