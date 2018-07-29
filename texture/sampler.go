package texture

import "image/color"

type ISampler interface {
	Sample(x, y float64, scaledHeight uint) color.NRGBA
}
