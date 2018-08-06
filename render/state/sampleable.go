package state

import (
	"image/color"

	"github.com/tlyakhov/gofoom/concepts"
)

type Sampleable interface {
	Sample(u, v float64, light *concepts.Vector3, scale float64) color.NRGBA
}
