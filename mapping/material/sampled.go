package material

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/texture"
)

type Sampled struct {
	*concepts.Base
	Sampler  texture.ISampler `editable:"Texture" edit_type:"Texture"`
	IsLiquid bool             `editable:"Is Liquid?" edit_type:"bool"`
}
