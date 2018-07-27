package engine

import (
	"math"

	"github.com/tlyakhov/gofoom/constants"

	"github.com/tlyakhov/gofoom/util"
)

type MapSector struct {
	util.CommonFields

	Segments                      []*MapSegment
	Entities                      map[string]*Entity
	Map                           *Map
	BottomZ, TopZ                 float64
	Min, Max, Center              *util.Vector3
	LightmapWidth, LightmapHeight uint
	FloorLightmap, CeilLightmap   []float64
	FloorScale, CeilScale         float64
	Hurt                          float64
	FloorTarget, CeilTarget       *MapSector
	// RoomImpulse
	// PVS
	// PVSEntity
	// PVSLights []
}

func (ms *MapSector) Update() {
	ms.Center = &util.Vector3{0, 0, (ms.TopZ + ms.BottomZ) / 2}
	ms.Min = &util.Vector3{math.Inf(1), math.Inf(1), ms.BottomZ}
	ms.Max = &util.Vector3{math.Inf(-1), math.Inf(-1), ms.TopZ}

	w := ms.Winding()

	for i, segment := range ms.Segments {
		next := ms.Segments[(i+1)%len(ms.Segments)]
		ms.Center.X += segment.A.X
		ms.Center.Y += segment.A.Y
		if segment.A.X < ms.Min.X {
			ms.Min.X = segment.A.X
		}
		if segment.A.Y < ms.Min.Y {
			ms.Min.Y = segment.A.Y
		}
		if segment.A.X > ms.Max.X {
			ms.Max.X = segment.A.X
		}
		if segment.A.Y > ms.Max.Y {
			ms.Max.Y = segment.A.Y
		}
		segment.Sector = ms
		segment.B = next.A
		segment.Update()

		if !w {
			segment.Normal = segment.Normal.Mul(-1)
		}
	}

	ms.Center = ms.Center.Mul(1.0 / float64(len(ms.Segments)))

	for _, e := range ms.Entities {
		e.Map = ms.Map
		e.Sector = ms
		e.Collide()
	}

	ms.LightmapWidth = uint((ms.Max.X-ms.Min.X)/constants.LightGrid) + 6
	ms.LightmapHeight = uint((ms.Max.Y-ms.Min.Y)/constants.LightGrid) + 6
	ms.FloorLightmap = make([]float64, ms.LightmapWidth*ms.LightmapHeight*3)
	ms.CeilLightmap = make([]float64, ms.LightmapWidth*ms.LightmapHeight*3)
	ms.ClearLightmaps()
}

func (ms *MapSector) ClearLightmaps() {
	for i := range ms.FloorLightmap {
		ms.FloorLightmap[i] = -1
		ms.CeilLightmap[i] = -1
	}

	for _, segment := range ms.Segments {
		segment.ClearLightmap()
	}

	// ms.UpdatePVS()
	// ms.UpdateEntityPVS()
}

func (ms *MapSector) IsPointInside2D(p *util.Vector2) bool {
	inside := false
	flag1 := (p.Y >= ms.Segments[0].A.Y)

	for i, segment := range ms.Segments {
		next := ms.Segments[(i+1)%len(ms.Segments)]
		flag2 := (p.Y >= next.A.Y)
		if flag1 != flag2 {
			if ((next.A.Y-p.Y)*(segment.A.X-next.A.X) >= (next.A.X-p.X)*(segment.A.Y-next.A.Y)) == flag2 {
				inside = !inside
			}
		}
		flag1 = flag2
	}
	return inside
}

func (ms *MapSector) Winding() bool {
	sum := 0.0
	for i, segment := range ms.Segments {
		next := ms.Segments[(i+1)%len(ms.Segments)]
		sum += (next.A.X - segment.A.X) * (segment.A.Y + next.A.Y)
	}
	return sum < 0
}

func (ms *MapSector) OnEnter(e *Entity) {
	if ms.FloorTarget == nil && e.Pos.Z <= e.Sector.BottomZ {
		e.Pos.Z = e.Sector.BottomZ
	}
}

func (ms *MapSector) OnExit(e *Entity) {
}
