package sectors

import (
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/registry"
)

type Underwater struct {
	core.PhysicalSector `editable:"^"`
}

func init() {
	registry.Instance().Register(Underwater{})
}
