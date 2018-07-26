package engine

import (
	"github.com/tlyakhov/gofoom/util"
)

type Ray struct {
	Start, End *util.Vector3
}

type RenderSlice struct {
	RenderTarget       []uint8
	X, Y, YStart, YEnd int
	TargetX            int
	Sector             *MapSector
	Segment            *MapSegment
	Ray                Ray
	RayIndex           int
	// Intersection
	Distance float64
	U        float64
	Depth    int
	Renderer *Renderer
}
