package material

import (
	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/math"
)

type Lit struct {
	concepts.Base
	Ambient  *math.Vector3 `editable:"Ambient Color" edit_type:"vector"`
	Diffuse  *math.Vector3 `editable:"Diffuse Color" edit_type:"vector"`
	IsLiquid bool          `editable:"Is Liquid?" edit_type:"bool"`
}
