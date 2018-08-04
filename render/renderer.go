package render

import (
	"fmt"
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/mapping"
)

type Renderer struct {
	*Config
	Map *mapping.Map
}

func NewRenderer() *Renderer {
	r := Renderer{
		Config: &Config{
			ScreenWidth:  640,
			ScreenHeight: 360,
			FOV:          constants.FieldOfView,
			MaxViewDist:  constants.MaxViewDistance,
			Frame:        0,
			FrameTint:    0,
			WorkerWidth:  640,
			Counter:      0,

			FloorNormal:   concepts.Vector3{X: 0, Y: 0, Z: 1},
			CeilingNormal: concepts.Vector3{X: 0, Y: 0, Z: -1},
		},
	}
	r.Initialize()
	return &r
}

func (r *Renderer) RenderSlice(slice *Slice) {
	slice.CalcScreen()

	slice.RenderCeiling()
	slice.RenderFloor()

	if slice.Segment.AdjacentSector == nil {
		slice.RenderMid()
		return
	}
	rsp := &SlicePortal{Slice: slice}
	rsp.CalcScreen()
	rsp.RenderHigh()
	rsp.RenderLow()

	portalSlice := *slice
	portalSlice.Sector = rsp.Adj
	portalSlice.YStart = rsp.AdjClippedTop
	portalSlice.YEnd = rsp.AdjClippedBottom
	portalSlice.Depth++
	r.RenderSector(&portalSlice)
}

func (r *Renderer) RenderSector(slice *Slice) {
	slice.Distance = constants.MaxViewDistance

	dist := math.MaxFloat64

	for _, segment := range slice.Sector.Segments {
		if slice.Ray.End.Sub(slice.Ray.Start).Dot(segment.Normal) > 0 {
			continue
		}

		isect := segment.Intersect(slice.Ray.Start, slice.Ray.End)

		if isect == nil {
			continue
		}

		delta := &concepts.Vector2{math.Abs(isect.X - slice.Ray.Start.X), math.Abs(isect.Y - slice.Ray.Start.Y)}
		if delta.Y > delta.X {
			dist = math.Abs(delta.Y / slice.AngleSin)
		} else {
			dist = math.Abs(delta.X / slice.AngleCos)
		}

		if dist > slice.Distance {
			continue
		}

		slice.Segment = segment
		slice.Distance = dist
		slice.Intersection = isect.To3D()
		slice.U = isect.Dist(segment.A) / segment.Length
	}

	if dist != math.MaxFloat64 {
		r.RenderSlice(slice)
	} else {
		fmt.Println("Depth: %v, sector: %s", slice.Depth, slice.Sector.ID)
	}
}

// Render a frame.
func (r *Renderer) Render(buffer []uint8) {
	r.Frame += 1
	r.Counter = 0
	xStart := 0
	xEnd := xStart + r.WorkerWidth

	for x := xStart; x < xEnd; x++ {
		// Reset the z-buffer to maximum viewing distance.
		for i := x - xStart; i < r.ScreenHeight*r.WorkerWidth+x-xStart; i += r.WorkerWidth {
			r.ZBuffer[i] = r.MaxViewDist
		}

		if r.Map.Player.Sector == nil {
			continue
		}

		// Initialize a slice...
		slice := &Slice{
			Config:       r.Config,
			Map:          r.Map,
			RenderTarget: buffer,
			X:            x,
			TargetX:      x - xStart,
			YStart:       0,
			YEnd:         r.ScreenHeight - 1,
			Angle:        r.Map.Player.Angle*concepts.Deg2rad + r.ViewRadians[x],
			Sector:       r.Map.Player.Sector.(*mapping.Sector),
			CameraZ:      r.Map.Player.Pos.Z + r.Map.Player.Height,
		}
		slice.AngleCos = math.Cos(slice.Angle)
		slice.AngleSin = math.Sin(slice.Angle)

		slice.Ray = &Ray{
			Start: r.Map.Player.Pos.To2D(),
			End: &concepts.Vector2{
				X: r.Map.Player.Pos.X + r.MaxViewDist*slice.AngleCos,
				Y: r.Map.Player.Pos.Y + r.MaxViewDist*slice.AngleSin,
			},
		}

		r.RenderSector(slice)
	}

	// Entities...
}
