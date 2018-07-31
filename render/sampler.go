package render

import (
	"image/color"

	"github.com/tlyakhov/gofoom/concepts"
)

type ISampler interface {
	Sample(slice *Slice, u, v float64, light *concepts.Vector3, scale float64) color.NRGBA
}
