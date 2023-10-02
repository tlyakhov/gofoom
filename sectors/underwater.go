package sectors

import (
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/registry"
)

type Underwater struct {
	core.PhysicalSector `editable:"^"`
}

func init() {
	registry.Instance().Register(Underwater{})
}

func (s *Underwater) Serialize() map[string]interface{} {
	result := s.PhysicalSector.Serialize()
	result["Type"] = "sectors.Underwater"
	return result
}
