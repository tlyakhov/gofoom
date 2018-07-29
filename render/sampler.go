package render

import (
	"image/color"

	"github.com/tlyakhov/gofoom/math"
)

type ISampler interface {
	Sample(slice *Slice, u, v float64, light *math.Vector3, scaledHeight uint) color.NRGBA
}
