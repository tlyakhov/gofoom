package materials

import (
	"fmt"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/registry"
)

type LitSampled struct {
	Lit
	Sampled
}

func init() {
	registry.Instance().Register(LitSampled{})
	registry.Instance().Register(PainfulLitSampled{})
}

func (m *LitSampled) SetParent(interface{}) {}

func (m *LitSampled) GetBase() *concepts.Base {
	return m.Lit.GetBase()
}

func (m *LitSampled) Initialize() {
	m.Sampled.Initialize()
	m.Lit.Initialize()
}

func (m *LitSampled) Deserialize(data map[string]interface{}) {
	m.Initialize()
	m.Lit.Deserialize(data)
	m.Sampled.Deserialize(data)
}

func (m *LitSampled) Serialize() map[string]interface{} {
	result := m.Lit.Serialize()
	result2 := m.Sampled.Serialize()
	for k, v := range result2 {
		result[k] = v
	}
	result["Type"] = "materials.LitSampled"
	return result
}

type PainfulLitSampled struct {
	LitSampled
	Painful
}

func (m *PainfulLitSampled) SetParent(interface{}) {}

func (m *PainfulLitSampled) GetBase() *concepts.Base {
	return m.LitSampled.GetBase()
}

func (m *PainfulLitSampled) Initialize() {
	m.LitSampled = LitSampled{}
	m.LitSampled.Initialize()
	m.Painful = Painful{}
	m.Painful.Initialize()
}

func (m *PainfulLitSampled) Deserialize(data map[string]interface{}) {
	m.Initialize()
	m.LitSampled.Deserialize(data)
	m.Painful.Deserialize(data)
	fmt.Printf("PainfulLitSampled: %v\n", m.Lit.GetBase().ID)
}

func (m *PainfulLitSampled) Serialize() map[string]interface{} {
	result := m.LitSampled.Serialize()
	result2 := m.Painful.Serialize()
	for k, v := range result2 {
		result[k] = v
	}
	result["Type"] = "materials.PainfulLitSampled"
	return result
}
