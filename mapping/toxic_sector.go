package mapping

import "github.com/tlyakhov/gofoom/registry"

type ToxicSector struct {
	Sector
	Hurt float64
}

func init() {
	registry.Instance().Register(ToxicSector{})
}

func (s *ToxicSector) Initialize() {
	s.Sector.Initialize()
}

func (s *ToxicSector) Deserialize(data map[string]interface{}) {
	s.Sector.Deserialize(data)

	if v, ok := data["Hurt"]; ok {
		s.Hurt = v.(float64)
	}
}
