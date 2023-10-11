package state

import (
	"tlyakhov/gofoom/concepts"
)

type Sampleable interface {
	Sample(u, v float64, light *concepts.Vector3, scale float64) uint32
}
