package material

import (
	"github.com/tlyakhov/gofoom/concepts"
)

type Sky struct {
	*concepts.Base
	Sampled
	StaticBackground bool `editable:"Static Background?" edit_type:"bool"`
}
