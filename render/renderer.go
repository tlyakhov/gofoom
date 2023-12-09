package render

import (
	"fmt"
	"math"
	"sync"

	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/render/state"
)

// Renderer holds all state related to a specific camera/map configuration.
type Renderer struct {
	*state.Config
	columnGroup *sync.WaitGroup
}

// NewRenderer constructs a new Renderer.
func NewRenderer(db *concepts.EntityComponentDB) *Renderer {
	r := Renderer{
		Config: &state.Config{
			ScreenWidth:  640,
			ScreenHeight: 360,
			FOV:          constants.FieldOfView,
			MaxViewDist:  constants.MaxViewDistance,
			Frame:        0,
			Counter:      0,
			DB:           db,
		},
		columnGroup: new(sync.WaitGroup),
	}

	r.Initialize()
	return &r
}

func (r *Renderer) RenderPortal(slice *state.Column) {
	if slice.Depth > constants.MaxPortals {
		dbg := fmt.Sprintf("Maximum portal depth reached @ %v", slice.Sector.Entity)
		slice.DebugNotices.Push(dbg)
		return
	}
	sp := &state.ColumnPortal{Column: slice}
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
	portalSlice.LightElements[0].Column = &portalSlice
	portalSlice.LightElements[1].Column = &portalSlice
	portalSlice.LightElements[2].Column = &portalSlice
	portalSlice.LightElements[3].Column = &portalSlice
	portalSlice.Sector = sp.Adj
	portalSlice.YStart = sp.AdjClippedTop
	portalSlice.YEnd = sp.AdjClippedBottom
	portalSlice.LastPortalDistance = slice.Distance
	portalSlice.Depth++
	r.RenderSector(&portalSlice)
	if slice.Pick {
		slice.PickedElements = portalSlice.PickedElements
	}
}

// RenderSegmentColumn draws or picks a single pixel vertical column given a particular
// segment intersection.
func (r *Renderer) RenderSegmentColumn(column *state.Column) {
	column.CalcScreen()
	column.Normal = column.Sector.CeilNormal
	if column.Pick {
		CeilingPick(column)
	} else {
		Ceiling(column)
	}
	column.Normal = column.Sector.FloorNormal
	if column.Pick {
		FloorPick(column)
	} else {
		Floor(column)
	}

	column.Segment.Normal.To3D(&column.Normal)

	hasPortal := !column.Segment.AdjacentSector.Nil()
	if column.Pick {
		if !hasPortal || column.Segment.PortalHasMaterial {
			WallMidPick(column)
			return
		}
		r.RenderPortal(column)
	} else {
		if hasPortal {
			r.RenderPortal(column)
		}
		if !hasPortal || column.Segment.PortalHasMaterial {
			WallMid(column)
		}
	}

}

// RenderSector intersects a camera ray for a single pixel column with a map sector.
func (r *Renderer) RenderSector(column *state.Column) {
	column.Distance = constants.MaxViewDistance

	dist := math.MaxFloat64
	isect := new(concepts.Vector2)
	ray := new(concepts.Vector2)
	ray.From(&column.Ray.End).SubSelf(&column.Ray.Start)
	for _, segment := range column.Sector.Segments {
		// Wall is facing away from us
		if ray.Dot(&segment.Normal) > 0 {
			continue
		}

		// Ray intersects?
		if ok := segment.Intersect2D(&column.Ray.Start, &column.Ray.End, isect); !ok {
			continue
		}

		delta := concepts.Vector2{math.Abs(isect[0] - column.Ray.Start[0]), math.Abs(isect[1] - column.Ray.Start[1])}
		if delta[1] > delta[0] {
			dist = math.Abs(delta[1] / column.AngleSin)
		} else {
			dist = math.Abs(delta[0] / column.AngleCos)
		}

		if dist > column.Distance || dist < column.LastPortalDistance {
			continue
		}

		column.Segment = segment
		column.Distance = dist
		isect.To3D(&column.Intersection)
		column.U = isect.Dist(&segment.P) / segment.Length
	}

	if dist != math.MaxFloat64 {
		r.RenderSegmentColumn(column)
	} else {
		dbg := fmt.Sprintf("No intersections for sector %v at depth: %v", column.Sector.Entity, column.Depth)
		r.DebugNotices.Push(dbg)
	}

	for _, ref := range column.Sector.Bodies {
		r.RenderBody(ref, column)
	}
}

// RenderColumn draws a single pixel column to an 8bit RGBA buffer.
func (r *Renderer) RenderColumn(column *state.Column, x int, y int, pick bool) []state.PickedElement {
	// Reset the z-buffer to maximum viewing distance.
	for i := x; i < r.ScreenHeight*r.ScreenWidth+x; i += r.ScreenWidth {
		r.ZBuffer[i] = r.MaxViewDist
	}

	// Reset the column
	column.Pick = pick
	column.X = x
	column.Y = y
	column.Angle = r.PlayerBody.Angle*concepts.Deg2rad + r.ViewRadians[x]
	column.Sector = r.PlayerBody.Sector()
	column.AngleCos = math.Cos(column.Angle)
	column.AngleSin = math.Sin(column.Angle)
	column.Ray.End = concepts.Vector2{
		r.PlayerBody.Pos.Render[0] + r.MaxViewDist*column.AngleCos,
		r.PlayerBody.Pos.Render[1] + r.MaxViewDist*column.AngleSin,
	}

	r.RenderSector(column)
	return column.PickedElements
}

func (r *Renderer) RenderBlock(buffer []uint8, xStart, xEnd int) {
	bob := math.Sin(r.Player.Bob)
	// Initialize a column...
	column := &state.Column{
		Config:  r.Config,
		YStart:  0,
		YEnd:    r.ScreenHeight,
		CameraZ: r.PlayerBody.Pos.Render[2] + r.PlayerBody.Height + bob,
	}
	column.LightElements[0].Column = column
	column.LightElements[1].Column = column
	column.LightElements[2].Column = column
	column.LightElements[3].Column = column

	column.Ray = &state.Ray{Start: *r.PlayerBody.Pos.Render.To2D()}

	for x := xStart; x < xEnd; x++ {
		if x >= xEnd {
			break
		}
		r.RenderColumn(column, x, 0, false)
		for y := 0; y < r.ScreenHeight; y++ {
			screenIndex := (x + y*r.ScreenWidth)
			fb := &r.FrameBuffer[screenIndex]
			screenIndex *= 4
			if r.FrameTint[3] != 0 {
				a := 1.0 - r.FrameTint[3]
				buffer[screenIndex+0] = uint8(concepts.Clamp((fb[0]*a+r.FrameTint[0])*255, 0, 255))
				buffer[screenIndex+1] = uint8(concepts.Clamp((fb[1]*a+r.FrameTint[1])*255, 0, 255))
				buffer[screenIndex+2] = uint8(concepts.Clamp((fb[2]*a+r.FrameTint[2])*255, 0, 255))
			} else {
				buffer[screenIndex+0] = uint8(concepts.Clamp(fb[0]*255, 0, 255))
				buffer[screenIndex+1] = uint8(concepts.Clamp(fb[1]*255, 0, 255))
				buffer[screenIndex+2] = uint8(concepts.Clamp(fb[2]*255, 0, 255))
			}
			buffer[screenIndex+3] = 0xFF
		}
	}

	if constants.RenderMultiThreaded {
		r.columnGroup.Done()
	}
}

// Render a frame.
func (r *Renderer) Render(buffer []uint8) {
	r.RefreshPlayer()

	// Frame Tint precalculation
	r.FrameTint = r.Player.FrameTint
	r.FrameTint[0] *= r.FrameTint[3]
	r.FrameTint[1] *= r.FrameTint[3]
	r.FrameTint[2] *= r.FrameTint[3]

	// Make sure we don't have too many debug notices
	for r.DebugNotices.Length() > 30 {
		r.DebugNotices.Pop()
	}
	if r.PlayerBody.SectorEntityRef.Nil() {
		r.DebugNotices.Push("Player is not in a sector")
		return
	}
	r.RenderLock.Lock()
	defer r.RenderLock.Unlock()

	r.Frame++
	r.Counter = 0

	if constants.RenderMultiThreaded {
		blocks := 24
		blockSize := r.ScreenWidth / blocks
		r.columnGroup.Add(blocks)
		for x := 0; x < blocks; x++ {
			go r.RenderBlock(buffer, x*blockSize, x*blockSize+blockSize)
		}
		r.columnGroup.Wait()
	} else {
		r.RenderBlock(buffer, 0, r.ScreenWidth)
	}
	// Bodies...
}

func (r *Renderer) Pick(x, y int) []state.PickedElement {
	if x < 0 || y < 0 || x >= r.ScreenWidth || y >= r.ScreenHeight {
		return nil
	}
	bob := math.Sin(r.Player.Bob)
	// Initialize a column...
	column := &state.Column{
		Config:  r.Config,
		YStart:  0,
		YEnd:    r.ScreenHeight,
		CameraZ: r.PlayerBody.Pos.Render[2] + r.PlayerBody.Height + bob,
	}
	column.LightElements[0].Column = column
	column.LightElements[1].Column = column
	column.LightElements[2].Column = column
	column.LightElements[3].Column = column

	column.Ray = &state.Ray{Start: *r.PlayerBody.Pos.Render.To2D()}
	return r.RenderColumn(column, x, y, true)
}
