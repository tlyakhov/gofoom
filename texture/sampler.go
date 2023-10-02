package texture

import "tlyakhov/gofoom/concepts"

type ISampler interface {
	concepts.ISerializable
	Sample(x, y float64, scale float64) uint32
}
