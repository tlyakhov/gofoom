package render

import (
	"image/color"
	"math"

	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/material"
)

func (m *material.Sampled) Sample(slice *Slice, u, v float64, light *math.Vector3, scaledHeight uint) color.NRGBA {
	if m.IsLiquid {
		u += math.Cos(float64(slice.Frame)*constants.LiquidChurnSpeed*deg2rad) * constants.LiquidChurnSize
		v += math.Cos(float64(slice.Frame)*constants.LiquidChurnSpeed*deg2rad) * constants.LiquidChurnSize
	}

	if u < 0 {
		u = math.Floor(u) - u
	} else if u >= 1.0 {
		u -= math.Floor(u)
	}

	if v < 0 {
		v = math.Floor(v) - v
	} else if v >= 1.0 {
		v -= math.Floor(v)
	}

	return m.Sampler.Sample(u, v, scaledHeight)
}
