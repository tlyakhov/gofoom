package engine

import (
	"math"

	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/engine/mapping"
	"github.com/tlyakhov/gofoom/util"
)

const (
	deg2rad float64 = math.Pi / 180.0
	rad2deg float64 = 180.0 / math.Pi
)

type trigEntry struct {
	sin, cos float64
}

type Renderer struct {
	Map                                    *mapping.Map
	ScreenWidth, ScreenHeight              int
	Frame, FrameTint, WorkerWidth, Counter int
	MaxViewDist, FOV                       float64

	cameraToProjectionPlane float64
	trigCount               int
	trigTable               []trigEntry
	viewFix                 []float64
	zbuffer                 []float64
	floorNormal             util.Vector3
	ceilingNormal           util.Vector3
}

func NewRenderer() *Renderer {
	r := Renderer{
		ScreenWidth:  640,
		ScreenHeight: 360,
		FOV:          constants.FieldOfView,
		MaxViewDist:  constants.MaxViewDistance,
		Frame:        0,
		FrameTint:    0,
		WorkerWidth:  640,
		Counter:      0,

		floorNormal:   util.Vector3{X: 0, Y: 0, Z: 1},
		ceilingNormal: util.Vector3{X: 0, Y: 0, Z: -1},
	}
	r.initTables()
	return &r
}

func (r *Renderer) initTables() {
	r.cameraToProjectionPlane = (float64(r.ScreenWidth) / 2.0) / math.Tan(r.FOV*deg2rad/2.0)
	r.trigCount = int(float64(r.ScreenWidth) * 360.0 / r.FOV) // Quantize trig tables per-Pixel.
	r.trigTable = make([]trigEntry, r.trigCount)
	r.viewFix = make([]float64, r.ScreenWidth)

	for i := 0; i < r.trigCount; i++ {
		r.trigTable[i].sin = math.Sin(float64(i) * 2.0 * math.Pi / float64(r.trigCount))
		r.trigTable[i].cos = math.Cos(float64(i) * 2.0 * math.Pi / float64(r.trigCount))
	}

	for i := 0; i < r.ScreenWidth/2; i++ {
		r.viewFix[i] = r.cameraToProjectionPlane / r.trigTable[r.ScreenWidth/2-1-i].cos
		r.viewFix[(r.ScreenWidth-1)-i] = r.viewFix[i]
	}

	r.zbuffer = make([]float64, r.WorkerWidth*r.ScreenHeight)

}

func (r *Renderer) RenderSector(slice RenderSlice) {

}

func (r *Renderer) normRayIndex(index int) int {
	for ; index < 0; index += r.trigCount {
	}
	for ; index >= r.trigCount; index -= r.trigCount {
	}
	return index
}

// Render a frame.
func (r *Renderer) Render(buffer []uint) {
	r.Counter = 0
	xStart := 0
	xEnd := xStart + r.WorkerWidth

	for x := xStart; x < xEnd; x++ {
		// Reset the z-buffer to maximum viewing distance.
		for i := 0; i < r.ScreenHeight; i++ {
			r.zbuffer[i*r.WorkerWidth+x-xStart] = r.MaxViewDist
		}

		if r.Map.Player.Sector == nil {
			continue
		}

		// Initialize a slice...
		slice := RenderSlice{
			Renderer:     r,
			RenderTarget: buffer,
			X:            x,
			TargetX:      x - xStart,
			YStart:       0,
			YEnd:         r.ScreenHeight - 1,
			RayIndex:     r.normRayIndex(int(r.Map.Player.Angle*float64(r.trigCount)/360.0) + x - r.ScreenWidth/2 + 1),
			Sector:       r.Map.Player.Sector,
		}

		slice.Ray = Ray{
			Start: r.Map.Player.Pos,
			End: util.Vector3{
				X: r.Map.Player.Pos.X + r.MaxViewDist*r.trigTable[slice.RayIndex].cos,
				Y: r.Map.Player.Pos.Y + r.MaxViewDist*r.trigTable[slice.RayIndex].sin,
			},
		}

		r.RenderSector(slice)
	}

	// Entities...
}
