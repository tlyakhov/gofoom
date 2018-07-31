package material

import (
	"github.com/tlyakhov/gofoom/concepts"
)

type Lit struct {
	*concepts.Base
	Ambient *concepts.Vector3 `editable:"Ambient Color" edit_type:"vector"`
	Diffuse *concepts.Vector3 `editable:"Diffuse Color" edit_type:"vector"`
}
