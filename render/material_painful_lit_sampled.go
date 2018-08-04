package render

import (
	"image/color"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/mapping/material"
	"github.com/tlyakhov/gofoom/registry"
)

type PainfulLitSampled material.PainfulLitSampled

func init() {
	registry.Instance().RegisterMapped(PainfulLitSampled{}, material.PainfulLitSampled{})
}

func (m *PainfulLitSampled) Sample(slice *Slice, u, v float64, light *concepts.Vector3, scale float64) color.NRGBA {
	return registry.Translate(m.LitSampled, "render").(*LitSampled).Sample(slice, u, v, light, scale)
}
