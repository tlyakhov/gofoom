package render

import (
	"fmt"
	"math"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/entities"
	"tlyakhov/gofoom/render/state"
)

// Renderer holds all state related to a specific camera/map configuration.
type Renderer struct {
	*state.Config
	Map     *core.Map
	columns chan int
}

// NewRenderer constructs a new Renderer.
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

// Player is a convenience function to get the player this renderer links to.
func (r *Renderer) Player() *entities.Player {
	return r.Map.Player.(*entities.Player)
}

// RenderSlice draws a single pixel vertical column given a particular segment intersection.
func (r *Renderer) RenderSlice(slice *state.Slice) {
	slice.CalcScreen()
	slice.Normal = slice.PhysicalSector.CeilNormal
	if slice.Pick {
		CeilingPick(slice)
	} else {
		Ceiling(slice)
	}
	slice.Normal = slice.PhysicalSector.FloorNormal
	if slice.Pick {
		FloorPick(slice)
	} else {
		Floor(slice)
	}

	slice.Segment.Normal.To3D(&slice.Normal)

	if slice.Segment.AdjacentSector == nil {
		if slice.Pick {
			WallMidPick(slice)
		} else {
			WallMid(slice)
		}
		return
	}
	if slice.Depth > 100 {
		fmt.Printf("Max depth reached!\n")
		return
	}
	sp := &state.SlicePortal{Slice: slice}
	sp.CalcScreen()
	if sp.AdjSegment != nil {
		if slice.Pick {
			WallHiPick(sp)
			WallLowPick(sp)
		} else {
			WallHi(sp)
			WallLow(sp)
		}
	}

	portalSlice := *slice
	portalSlice.LightElements[0].Slice = &portalSlice
	portalSlice.LightElements[1].Slice = &portalSlice
	portalSlice.LightElements[2].Slice = &portalSlice
	portalSlice.LightElements[3].Slice = &portalSlice
	portalSlice.PhysicalSector = sp.Adj.Physical()
	portalSlice.YStart = sp.AdjClippedTop
	portalSlice.YEnd = sp.AdjClippedBottom
	portalSlice.LastPortalDistance = slice.Distance
	portalSlice.Depth++
	r.RenderSector(&portalSlice)
	if slice.Pick {
		slice.PickedElements = portalSlice.PickedElements
	}
}

// RenderSector intersects a camera ray for a single pixel column with a map sector.
func (r *Renderer) RenderSector(slice *state.Slice) {
	slice.Distance = constants.MaxViewDistance

	dist := math.MaxFloat64
	isect := &concepts.Vector2{}
	ray := &concepts.Vector2{}
	ray.From(&slice.Ray.End).SubSelf(&slice.Ray.Start)
	for _, segment := range slice.PhysicalSector.Segments {
		// Wall is facing away from us
		if ray.Dot(&segment.Normal) > 0 {
			continue
		}

		// Ray intersects?
		if ok := segment.Intersect2D(&slice.Ray.Start, &slice.Ray.End, isect); !ok {
			continue
		}

		delta := concepts.Vector2{math.Abs(isect[0] - slice.Ray.Start[0]), math.Abs(isect[1] - slice.Ray.Start[1])}
		if delta[1] > delta[0] {
			dist = math.Abs(delta[1] / slice.AngleSin)
		} else {
			dist = math.Abs(delta[0] / slice.AngleCos)
		}

		if dist > slice.Distance || dist < slice.LastPortalDistance {
			continue
		}

		slice.Segment = segment
		slice.Distance = dist
		isect.To3D(&slice.Intersection)
		slice.U = isect.Dist(&segment.P) / segment.Length
	}

	if dist != math.MaxFloat64 {
		r.RenderSlice(slice)
	} else {
		fmt.Printf("Depth: %v, sector: %s\n", slice.Depth, slice.PhysicalSector.ID)
	}
}

// RenderColumn draws a single pixel column to an 8bit RGBA buffer.
func (r *Renderer) RenderColumn(buffer []uint8, x int, y int, pick bool) []state.PickedElement {
	// Reset the z-buffer to maximum viewing distance.
	for i := x; i < r.ScreenHeight*r.ScreenWidth+x; i += r.ScreenWidth {
		r.ZBuffer[i] = r.MaxViewDist
	}

	bob := math.Sin(r.Player().Bob)
	// Initialize a slice...
	slice := &state.Slice{
		Config:         r.Config,
		Map:            r.Map,
		RenderTarget:   buffer,
		Pick:           pick,
		X:              x,
		Y:              y,
		YStart:         0,
		YEnd:           r.ScreenHeight,
		Angle:          r.Player().Angle*concepts.Deg2rad + r.ViewRadians[x],
		PhysicalSector: r.Player().Sector.Physical(),
		CameraZ:        r.Player().Pos[2] + r.Player().Height + bob,
	}
	slice.AngleCos = math.Cos(slice.Angle)
	slice.AngleSin = math.Sin(slice.Angle)
	slice.LightElements[0].Slice = slice
	slice.LightElements[1].Slice = slice
	slice.LightElements[2].Slice = slice
	slice.LightElements[3].Slice = slice

	slice.Ray = &state.Ray{
		Start: *r.Player().Pos.To2D(),
		End: concepts.Vector2{
			r.Player().Pos[0] + r.MaxViewDist*slice.AngleCos,
			r.Player().Pos[1] + r.MaxViewDist*slice.AngleSin,
		},
	}

	r.RenderSector(slice)
	return slice.PickedElements
}

func (r *Renderer) RenderBlock(buffer []uint8, xStart, xEnd int) {
	for x := xStart; x < xEnd; x++ {
		if x >= xEnd {
			break
		}
		r.RenderColumn(buffer, x, 0, false)
	}

	if constants.RenderMultiThreaded {
		r.columns <- xStart
	}
}

// Render a frame.
func (r *Renderer) Render(buffer []uint8) {
	if r.Player().Sector == nil {
		return
	}
	r.Map.RenderLock.Lock()
	defer r.Map.RenderLock.Unlock()

	r.Frame++
	r.Counter = 0

	if constants.RenderMultiThreaded {
		blockSize := r.ScreenWidth / 8
		blocks := 8
		for x := 0; x < blocks; x++ {
			go r.RenderBlock(buffer, x*blockSize, x*blockSize+blockSize)
		}

		for x := 0; x < blocks; x++ {
			<-r.columns
		}
	} else {
		for x := 0; x < r.ScreenWidth; x++ {
			r.RenderColumn(buffer, x, 0, false)
		}
	}
	// Entities...
}

func (r *Renderer) Pick(x, y int) []state.PickedElement {
	if x < 0 || y < 0 || x >= r.ScreenWidth || y >= r.ScreenHeight {
		return nil
	}
	return r.RenderColumn(nil, x, y, true)
}
