package engine

import (
	"github.com/tlyakhov/gofoom/util"
)

type MapSegment struct {
	util.CommonFields

	Ax, Ay, Bx, By float64
	// LoMaterial, HiMaterial, MidMaterial
	Length          float64
	Normal          util.Vector3
	Sector          *MapSector
	AdjacentSector  *MapSector
	AdjacentSegment *MapSegment
	Lightmap        []float64
	LightmapWidth   int
	LightmapHeight  int
	Flags           int
}
