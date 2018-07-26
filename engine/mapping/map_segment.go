package mapping

import (
	"github.com/tlyakhov/gofoom/util"
)

type MapSegment struct {
	Ax, Ay, Bx, By float64
	// LoMaterial, HiMaterial, MidMaterial
	Length         float64
	Normal         util.Vector3
	AdjacentSector *MapSector
}
