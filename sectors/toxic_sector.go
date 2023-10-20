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

func (s *ToxicSector) Construct(data map[string]interface{}) {
	s.PhysicalSector.Construct(data)
	s.Model = s

	if data == nil {
		return
	}

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
