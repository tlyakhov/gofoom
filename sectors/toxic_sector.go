package sectors

import (
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/registry"
)

type ToxicSector struct {
	core.PhysicalSector `editable:"^"`
	Hurt                float64 `editable:"Hurt Amount"`
}

func init() {
	registry.Instance().Register(ToxicSector{})
}

func (s *ToxicSector) Initialize() {
	s.PhysicalSector.Initialize()
}

func (s *ToxicSector) Deserialize(data map[string]interface{}) {
	s.PhysicalSector.Deserialize(data)

	if v, ok := data["Hurt"]; ok {
		s.Hurt = v.(float64)
	}
}

func (s *ToxicSector) Serialize() map[string]interface{} {
	result := s.PhysicalSector.Serialize()
	result["Type"] = "sectors.ToxicSector"
	result["Hurt"] = s.Hurt
	return result
}
