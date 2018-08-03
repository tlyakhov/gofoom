package render

import (
	"image/color"
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/mapping/material"
	"github.com/tlyakhov/gofoom/registry"
)

type Sampled material.Sampled

func init() {
	registry.Instance().RegisterMapped(Sampled{}, material.Sampled{})
}

func (m *Sampled) Sample(slice *Slice, u, v float64, light *concepts.Vector3, scale float64) color.NRGBA {
	if m.IsLiquid {
		u += math.Cos(float64(slice.Frame)*constants.LiquidChurnSpeed*concepts.Deg2rad) * constants.LiquidChurnSize
		v += math.Cos(float64(slice.Frame)*constants.LiquidChurnSpeed*concepts.Deg2rad) * constants.LiquidChurnSize
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

	return m.Sampler.Sample(u, v, scale)
}
