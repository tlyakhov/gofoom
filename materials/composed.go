package materials

import (
	"fmt"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/registry"
)

type LitSampled struct {
	Lit     `editable:"^"`
	Sampled `editable:"^"`
}

func init() {
	registry.Instance().Register(LitSampled{})
	registry.Instance().Register(PainfulLitSampled{})
}

func (m *LitSampled) SetParent(interface{}) {}

func (m *LitSampled) GetBase() *concepts.Base {
	return m.Lit.GetBase()
}

func (m *LitSampled) GetModel() concepts.ISerializable {
	return m.Lit.Model
}

func (m *LitSampled) Construct(data map[string]interface{}) {
	m.Lit.Construct(data)
	m.Lit.Model = m
	m.Sampled.Construct(data)
	m.Sampled.Model = m
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

func (m *LitSampled) Sample(u, v float64, light *concepts.Vector3, scale float64) uint32 {
	surface := concepts.Int32ToVector3(m.Sampled.Sample(u, v, light, scale))
	amb := *light
	amb.AddSelf(&m.Ambient)
	sum := (&surface).Mul3Self(&m.Diffuse).Mul3Self(&amb).ClampSelf(0.0, 255.0)
	return sum.ToInt32Color()
}

type PainfulLitSampled struct {
	LitSampled `editable:"^"`
	Painful    `editable:"^"`
}

func (m *PainfulLitSampled) SetParent(interface{}) {}

func (m *PainfulLitSampled) GetBase() *concepts.Base {
	return m.LitSampled.GetBase()
}

func (m *PainfulLitSampled) Construct(data map[string]interface{}) {
	m.LitSampled.Construct(data)
	m.LitSampled.Lit.Model = m
	m.LitSampled.Sampled.Model = m
	m.Painful.Construct(data)
	m.Painful.Model = m
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
