package sectors

import (
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/registry"
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
