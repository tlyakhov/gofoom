package texture

import "image/color"

type ISampler interface {
	Sample(x, y float64, scale float64) color.NRGBA
}
