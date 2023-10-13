package core

import (
	"tlyakhov/gofoom/concepts"
)

type Sampleable interface {
	concepts.ISerializable
	Sample(u, v float64, light *concepts.Vector3, scale float64) uint32
}
