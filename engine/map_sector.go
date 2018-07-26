package engine

import (
	"github.com/tlyakhov/gofoom/util"
)

type MapSector struct {
	util.CommonFields

	Segments                      []MapSegment
	Entities                      []Entity
	Map                           *Map
	BottomZ, TopZ                 float64
	Min, Max, Center              util.Vector3
	LightmapWidth, LightmapHeight int
	FloorScale, CeilScale         float64
	Hurt                          float64
	FloorTarget, CeilTarget       *MapSector
	Version                       int
	// RoomImpulse
	// PVS
	// PVSEntity
	// PVSLights []
}
