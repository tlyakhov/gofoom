package render

import (
	"fmt"
	"math"

	"github.com/tlyakhov/gofoom/concepts"
	"github.com/tlyakhov/gofoom/constants"
	"github.com/tlyakhov/gofoom/core"
	"github.com/tlyakhov/gofoom/entities"
	"github.com/tlyakhov/gofoom/render/state"
)

type Renderer struct {
	*state.Config
	Map     *core.Map
	columns chan int
}

func NewRenderer() *Renderer {
	r := Renderer{
		Config: &state.Config{
			ScreenWidth:  640,
			ScreenHeight: 360,
			FOV:          constants.FieldOfView,
			MaxViewDist:  constants.MaxViewDistance,
			Frame:        0,
			Counter:      0,
		},
		columns: make(chan int),
	}
	r.Initialize()
	return &r
}

func (r *Renderer) Player() *entities.Player {
	return r.Map.Player.(*entities.Player)
}

func (r *Renderer) RenderSlice(slice *state.Slice) {
	slice.CalcScreen()
	Ceiling(slice)
	Floor(slice)

	if slice.Segment.AdjacentSector == nil {
		WallMid(slice)
		return
	}
	sp := &state.SlicePortal{Slice: slice}
	sp.CalcScreen()
	WallHi(sp)
	WallLow(sp)

	portalSlice := *slice
	portalSlice.PhysicalSector = sp.Adj.Physical()
	portalSlice.YStart = sp.AdjClippedTop
	portalSlice.YEnd = sp.AdjClippedBottom
	portalSlice.Depth++
	r.RenderSector(&portalSlice)
}

func (r *Renderer) RenderSector(slice *state.Slice) {
	slice.Distance = constants.MaxViewDistance

	dist := math.MaxFloat64

	for _, segment := range slice.PhysicalSector.Segments {
		if slice.Ray.End.Sub(slice.Ray.Start).Dot(segment.Normal) > 0 {
			continue
		}

		isect, ok := segment.Intersect2D(slice.Ray.Start, slice.Ray.End)

		if !ok {
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
		fmt.Printf("Depth: %v, sector: %s\n", slice.Depth, slice.PhysicalSector.ID)
	}
}

func (r *Renderer) RenderColumn(buffer []uint8, x int) {
	// Reset the z-buffer to maximum viewing distance.
	for i := x; i < r.ScreenHeight*r.ScreenWidth+x; i += r.ScreenWidth {
		r.ZBuffer[i] = r.MaxViewDist
	}

	// Initialize a slice...
	slice := &state.Slice{
		Config:         r.Config,
		Map:            r.Map,
		RenderTarget:   buffer,
		X:              x,
		YStart:         0,
		YEnd:           r.ScreenHeight,
		Angle:          r.Player().Angle*concepts.Deg2rad + r.ViewRadians[x],
		PhysicalSector: r.Player().Sector.Physical(),
		CameraZ:        r.Player().Pos.Z + r.Player().Height,
	}
	slice.AngleCos = math.Cos(slice.Angle)
	slice.AngleSin = math.Sin(slice.Angle)

	slice.Ray = &state.Ray{
		Start: r.Player().Pos.To2D(),
		End: concepts.Vector2{
			X: r.Player().Pos.X + r.MaxViewDist*slice.AngleCos,
			Y: r.Player().Pos.Y + r.MaxViewDist*slice.AngleSin,
		},
	}

	r.RenderSector(slice)

	r.columns <- x
}

// Render a frame.
func (r *Renderer) Render(buffer []uint8) {
	if r.Player().Sector == nil {
		return
	}

	r.Frame++
	r.Counter = 0

	for x := 0; x < r.ScreenWidth; x++ {
		go r.RenderColumn(buffer, x)
	}
	for x := 0; x < r.ScreenWidth; x++ {
		<-r.columns
	}
	// Entities...
}
