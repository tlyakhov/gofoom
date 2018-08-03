package mapping

import "github.com/tlyakhov/gofoom/registry"

type ToxicSector struct {
	Sector
	Hurt float64
}

func init() {
	registry.Instance().Register(ToxicSector{})
}
