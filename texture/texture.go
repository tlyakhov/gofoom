package texture

import (
	"image/color"

	// Decoders for common image types
	_ "image/jpeg"
	_ "image/png"

	"github.com/tlyakhov/gofoom/concepts"
)

type Solid struct {
	concepts.Base
	diffuse color.NRGBA
}

func (s *Solid) Sample(x, y float64, scaledHeight uint) color.NRGBA {
	return s.diffuse
}
