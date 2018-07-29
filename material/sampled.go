package material

import (
	"github.com/tlyakhov/gofoom/texture"
)

type Sampled struct {
	Sampler *texture.ISampler `editable:"Texture" edit_type:"Texture"`
}

func (m *Sampled) ActOnEntity(e Entity) {
}
