package engine

import (
	"image/color"
	"math"

	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/util"
)

type Material struct {
	util.CommonFields

	Texture          *Texture      `editable:"Texture" edit_type:"Texture"`
	Ambient          *util.Vector3 `editable:"Ambient Color" edit_type:"vector"`
	Diffuse          *util.Vector3 `editable:"Diffuse Color" edit_type:"vector"`
	RenderAsSky      bool          `editable:"Render as sky?" edit_type:"bool"`
	StaticBackground bool          `editable:"Static Background?" edit_type:"bool"`
	Hurt             int           `editable:"Hurt" edit_type:"float"`
	IsLiquid         bool          `editable:"Is Liquid?" edit_type:"bool"`

	// Runtime references
	Map *Map
}

func (m *Material) Sample(slice *RenderSlice, u, v float64, light *util.Vector3, scaledHeight uint) color.NRGBA {
	if m.RenderAsSky {
		v = float64(slice.Y) / (float64(slice.Renderer.ScreenHeight) - 1)

		if m.StaticBackground {
			u = float64(slice.X) / (float64(slice.Renderer.ScreenWidth) - 1)
		} else {
			u = float64(slice.RayIndex) / (float64(slice.Renderer.trigCount) - 1)
		}
		// Assume largest mipmap
		scaledHeight = 0
	}

	if m.IsLiquid {
		u += math.Cos(float64(slice.Renderer.Frame)*constants.LiquidChurnSpeed*deg2rad) * constants.LiquidChurnSize
		v += math.Cos(float64(slice.Renderer.Frame)*constants.LiquidChurnSpeed*deg2rad) * constants.LiquidChurnSize
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

	surface := m.Texture.Sample(u, v, scaledHeight)

	if m.RenderAsSky {
		return surface
	}
	sum := &util.Vector3{float64(surface.R), float64(surface.G), float64(surface.B)}
	sum = sum.Mul3(m.Diffuse)

	if light != nil {
		sum = sum.Mul3(light)
	}
	sum = sum.Add(m.Ambient).Clamp(0.0, 255.0)
	return color.NRGBA{uint8(sum.X), uint8(sum.Y), uint8(sum.Z), 255}
}
