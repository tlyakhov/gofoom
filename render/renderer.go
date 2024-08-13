// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"fmt"
	"math"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/puzpuzpuz/xsync/v3"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/render/state"
)

// Renderer holds all state related to a specific camera/map configuration.
type Renderer struct {
	*state.Config
	Columns            []state.Column
	columnGroup        *sync.WaitGroup
	startingSector     *core.Sector
	textStyle          *TextStyle
	SectorLastRendered *xsync.MapOf[concepts.Entity, uint64]

	ICacheHits, ICacheMisses atomic.Int64
}

// NewRenderer constructs a new Renderer.
func NewRenderer(db *concepts.EntityComponentDB) *Renderer {
	r := Renderer{
		Config: &state.Config{
			ScreenWidth:   640,
			ScreenHeight:  360,
			FOV:           constants.FieldOfView,
			Multithreaded: constants.RenderMultiThreaded,
			Blocks:        constants.RenderBlocks,
			LightGrid:     constants.LightGrid,
			MaxViewDist:   constants.MaxViewDistance,
			DB:            db,
		},
		columnGroup:        new(sync.WaitGroup),
		SectorLastRendered: xsync.NewMapOf[concepts.Entity, uint64](),
	}

	r.Initialize()
	return &r
}

func (r *Renderer) Initialize() {
	r.Config.Initialize()

	r.Columns = make([]state.Column, r.Blocks)

	for i := range r.Columns {
		r.Columns[i].Config = r.Config
		r.Columns[i].PortalColumns = make([]state.Column, constants.MaxPortals)
		r.Columns[i].Visited = make([]state.SegmentIntersection, constants.MaxPortals)
		// Set up 16 slots initially
		r.Columns[i].EntitiesByDistance = make([]state.EntityWithDist2, 0, 16)
		r.Columns[i].LightLastColIndices = make([]uint64, r.ScreenHeight)
		r.Columns[i].LightLastColResults = make([]concepts.Vector3, r.ScreenHeight*8)
	}
	r.textStyle = r.NewTextStyle()
}

func (r *Renderer) WorldToScreen(world *concepts.Vector3) *concepts.Vector2 {
	relative := concepts.Vector2{world[0], world[1]}
	relative[0] -= r.PlayerBody.Pos.Render[0]
	relative[1] -= r.PlayerBody.Pos.Render[1]
	radians := math.Atan2(relative[1], relative[0]) - *r.PlayerBody.Angle.Render*concepts.Deg2rad
	for radians < -math.Pi {
		radians += 2 * math.Pi
	}
	if radians < -math.Pi*0.5 || radians > math.Pi*0.5 {
		return nil
	}
	x := math.Tan(radians)*r.CameraToProjectionPlane + float64(r.ScreenWidth)*0.5
	dist := relative.Length()
	y := (world[2] - r.Player.CameraZ) / dist
	y *= r.CameraToProjectionPlane / math.Cos(radians)
	y = float64(r.ScreenHeight/2) - math.Floor(y)
	return &concepts.Vector2{x, y}
}

func (r *Renderer) RenderPortal(c *state.Column) {
	if c.Depth >= constants.MaxPortals-1 {
		dbg := fmt.Sprintf("Maximum portal depth reached @ %v", c.Sector.Entity)
		c.DebugNotices.Push(dbg)
		return
	}

	// Copy current column into portal column
	next := &c.PortalColumns[c.Depth]
	*next = *c

	if c.SectorSegment.AdjacentSegment.PortalTeleports {
		next.Ray = &state.Ray{Start: c.Start, End: c.End}
		next.Ray.Start = *c.SectorSegment.PortalMatrix.Unproject(&next.Ray.Start)
		next.Ray.End = *c.SectorSegment.PortalMatrix.Unproject(&next.Ray.End)
		next.Ray.Start = *c.SectorSegment.AdjacentSegment.MirrorPortalMatrix.Project(&next.Ray.Start)
		next.Ray.End = *c.SectorSegment.AdjacentSegment.MirrorPortalMatrix.Project(&next.Ray.End)
		next.Ray.AnglesFromStartEnd()
		// TODO: this has a bug if the adjacent sector has a sloped floor.
		// Getting the right floor height is a bit expensive because we have to
		// project the intersection point. For now just use the sector minimum.
		next.CameraZ = c.CameraZ - c.IntersectionBottom + c.SectorSegment.AdjacentSegment.Sector.Min[2]
		next.RayFloorCeil[0] = next.Ray.AngleCos * c.ViewFix[c.ScreenX]
		next.RayFloorCeil[1] = next.Ray.AngleSin * c.ViewFix[c.ScreenX]
		next.MaterialSampler.Ray = next.Ray
	}

	// This allocation is ok, does not escape
	portal := &state.ColumnPortal{Column: next}
	portal.CalcScreen()
	if portal.AdjSegment != nil {
		if c.Pick {
			wallHiPick(portal)
			wallLowPick(portal)
		} else {
			wallHi(portal)
			wallLow(portal)
		}
	}

	next.Sector = portal.Adj
	next.EdgeTop = portal.AdjClippedTop
	next.EdgeBottom = portal.AdjClippedBottom
	next.LastPortalDistance = c.Distance
	next.Depth++
	r.RenderSector(next)
	c.PickedSelection = next.PickedSelection
}

// RenderSegmentColumn draws or picks a single pixel vertical column given a particular
// segment intersection.
func (r *Renderer) RenderSegmentColumn(c *state.Column) {
	c.CalcScreen()

	c.LightSampler.MaterialSampler.Config = r.Config
	c.LightSampler.Type = state.LightSamplerCeil
	c.LightSampler.Normal = c.Sector.CeilNormal
	c.LightSampler.Sector = c.Sector
	c.LightSampler.Segment = c.Segment

	if c.Pick {
		ceilingPick(c)
	} else {
		ceiling(c)
	}
	c.LightSampler.Type = state.LightSamplerFloor
	c.LightSampler.Normal = c.Sector.FloorNormal

	if c.Pick {
		floorPick(c)
	} else {
		floor(c)
	}

	c.LightSampler.Type = state.LightSamplerWall
	c.Segment.Normal.To3D(&c.LightSampler.Normal)

	hasPortal := c.SectorSegment.AdjacentSector != 0 && c.SectorSegment.AdjacentSegment != nil
	if c.Pick {
		if !hasPortal || c.SectorSegment.PortalHasMaterial {
			wallPick(c)
			return
		}
		r.RenderPortal(c)
	} else {
		if hasPortal {
			r.RenderPortal(c)
		}
		if !hasPortal || c.SectorSegment.PortalHasMaterial {
			r.wall(c)
		}
	}

}

// RenderSector intersects a camera ray for a single pixel column with a map sector.
func (r *Renderer) RenderSector(c *state.Column) {
	// Remember the frame # we rendered this sector. This is used when trying to
	// invalidate lighting caches (Sector.Lightmap)
	// TODO: We can probably do something simpler and cheaper here.
	r.SectorLastRendered.Store(c.Sector.Entity, c.DB.Frame)

	if c.Sector.LightmapBias[0] == math.MaxInt64 {
		// Floor is important, needs to truncate towards -Infinity rather than 0
		c.Sector.LightmapBias[0] = int64(math.Floor(c.Sector.Min[0]/r.LightGrid)) - 2
		c.Sector.LightmapBias[1] = int64(math.Floor(c.Sector.Min[1]/r.LightGrid)) - 2
		c.Sector.LightmapBias[2] = int64(math.Floor(c.Sector.Min[2]/r.LightGrid)) - 2
	}

	/*  The structure of this function is a bit complicated because we try to
		remember successful ray/segment intersections for convex sectors. This
		is most beneficial for sectors with a lot of segments.

		Summary of Renderer.RenderSector overall:

		1. If the sector is not concave, and the contents of the cache at the
		    current portal depth match the current sector, and the cached segment
	        is intersected by the ray, use the cached data, avoiding visiting the
		    rest of the segments for the sector.

		2. Otherwise, iterate through all the segments looking for
		   intersections.

		3. Render a column, potentially visiting portal sectors.
		4. Find intersections with internal segments, collect distances.
		5. Find intersections with Bodies, collect distances.
		6. Sort bodies/internal segments by distance.
		7. Render bodies and internal segments in order.

	*/

	c.SegmentIntersection = &c.Visited[c.Depth]
	cacheValid := !c.Sector.Concave && c.SectorSegment != nil && c.SectorSegment.Sector == c.Sector
	if cacheValid && c.SectorSegment.Intersect2D(&c.Ray.Start, &c.Ray.End, &c.RaySegTest) {
		r.ICacheHits.Add(1)
		c.Distance = c.Ray.DistTo(&c.RaySegTest)
		c.RaySegIntersect[0] = c.RaySegTest[0]
		c.RaySegIntersect[1] = c.RaySegTest[1]
	} else {
		r.ICacheMisses.Add(1)
		found := false
		for _, sectorSeg := range c.Sector.Segments {
			// Wall is facing away from us
			if c.Ray.Delta.Dot(&sectorSeg.Normal) > 0 {
				continue
			}

			// Ray intersects?
			if !sectorSeg.Intersect2D(&c.Ray.Start, &c.Ray.End, &c.RaySegTest) {
				continue
			}

			// Check if we've already found a closer segment
			dist := c.Ray.DistTo(&c.RaySegTest)
			if (found && dist > c.Distance) ||
				dist < c.LastPortalDistance {
				continue
			}

			found = true
			c.Segment = &sectorSeg.Segment
			c.SectorSegment = sectorSeg
			c.Distance = dist
			c.RaySegIntersect[0] = c.RaySegTest[0]
			c.RaySegIntersect[1] = c.RaySegTest[1]
		}
		if !found {
			c.SegmentIntersection = nil
		}
	}

	if c.SegmentIntersection != nil {
		c.SegmentIntersection.U = c.RaySegIntersect.To2D().Dist(c.SectorSegment.A) / c.SectorSegment.Length
		c.IntersectionBottom, c.IntersectionTop = c.Sector.PointZ(concepts.DynamicRender, c.RaySegIntersect.To2D())
		r.RenderSegmentColumn(c)
	} else {
		dbg := fmt.Sprintf("No intersections for sector %v at depth: %v", c.Sector.Entity, c.Depth)
		r.DebugNotices.Push(dbg)
	}

	// Clear slice without reallocating memory
	c.EntitiesByDistance = c.EntitiesByDistance[:0]
	for _, b := range c.Sector.Bodies {
		c.EntitiesByDistance = append(c.EntitiesByDistance, state.EntityWithDist2{
			Body:  b,
			Dist2: c.Ray.Start.Dist2(b.Pos.Render.To2D()),
		})
	}
	for _, s := range c.Sector.InternalSegments {
		// TODO: we do this again later. Should we optimize this?
		if s == nil || !s.IsActive() {
			continue
		}
		// Is the segment facing away?
		if !s.TwoSided && c.Ray.Delta.Dot(&s.Normal) > 0 {
			continue
		}
		// Ray intersects?
		if !s.Intersect2D(&c.Ray.Start, &c.Ray.End, &c.RaySegTest) {
			continue
		}

		dist := c.Ray.DistTo(&c.RaySegTest)
		c.EntitiesByDistance = append(c.EntitiesByDistance, state.EntityWithDist2{
			InternalSegment: s,
			Dist2:           dist * dist,
		})
	}

	slices.SortFunc(c.EntitiesByDistance, func(a state.EntityWithDist2, b state.EntityWithDist2) int {
		return int(b.Dist2 - a.Dist2)
	})
	c.SegmentIntersection = &state.SegmentIntersection{}
	for _, sorted := range c.EntitiesByDistance {
		if sorted.Body != nil {
			r.renderBody(sorted.Body, c)
			continue
		}

		// Ray intersection
		sorted.InternalSegment.Intersect2D(&c.Ray.Start, &c.Ray.End, &c.RaySegTest)
		c.Segment = &sorted.InternalSegment.Segment
		c.Distance = c.Ray.DistTo(&c.RaySegTest)
		c.RaySegIntersect[0] = c.RaySegTest[0]
		c.RaySegIntersect[1] = c.RaySegTest[1]
		c.SegmentIntersection.U = c.RaySegTest.Dist(sorted.InternalSegment.A) / sorted.InternalSegment.Length
		c.IntersectionTop = sorted.InternalSegment.Top
		c.IntersectionBottom = sorted.InternalSegment.Bottom
		c.CalcScreen()

		if c.Pick && c.ScreenY >= c.ClippedTop && c.ScreenY <= c.ClippedBottom {
			c.PickedSelection = append(c.PickedSelection, core.SelectableFromInternalSegment(sorted.InternalSegment))
			return
		}
		c.LightSampler.MaterialSampler.Config = r.Config
		c.LightSampler.Sector = c.Sector
		c.LightSampler.Segment = &sorted.InternalSegment.Segment
		c.LightSampler.Type = state.LightSamplerWall
		sorted.InternalSegment.Normal.To3D(&c.LightSampler.Normal)
		r.wall(c)
	}
}

// RenderColumn draws a single pixel column to an 8bit RGBA buffer.
func (r *Renderer) RenderColumn(column *state.Column, x int, y int, pick bool) []*core.Selectable {
	// Reset the z-buffer to maximum viewing distance.
	for i := x; i < r.ScreenHeight*r.ScreenWidth+x; i += r.ScreenWidth {
		r.ZBuffer[i] = r.MaxViewDist
	}

	// Reset the column
	column.LastPortalDistance = 0
	column.Depth = 0
	column.EdgeTop = 0
	column.EdgeBottom = r.ScreenHeight
	column.Pick = pick
	column.ScreenX = x
	column.ScreenY = y
	column.MaterialSampler.ScreenX = x
	column.MaterialSampler.ScreenY = y
	column.MaterialSampler.Angle = column.Angle
	column.Ray.Set(*r.PlayerBody.Angle.Render*concepts.Deg2rad + r.ViewRadians[x])
	column.RayFloorCeil[0] = column.Ray.AngleCos * column.ViewFix[column.ScreenX]
	column.RayFloorCeil[1] = column.Ray.AngleSin * column.ViewFix[column.ScreenX]

	if r.startingSector != nil {
		column.Sector = r.startingSector
	}
	if column.Sector == nil {
		return nil
	}

	r.RenderSector(column)
	return column.PickedSelection
}

func (r *Renderer) RenderBlock(columnIndex, xStart, xEnd int) {
	// Initialize a column...
	column := &r.Columns[columnIndex]
	column.BodiesSeen = make(map[concepts.Entity]*core.Body)
	column.CameraZ = r.Player.CameraZ
	column.Ray = &state.Ray{Start: *r.PlayerBody.Pos.Render.To2D()}
	column.MaterialSampler = state.MaterialSampler{Config: r.Config, Ray: column.Ray}
	column.LightSampler.XorSeed = r.DB.Frame + uint64(xStart)
	for i := range column.LightLastColIndices {
		column.LightLastColIndices[i] = 0
	}

	for x := xStart; x < xEnd; x++ {
		if x >= r.ScreenWidth {
			break
		}
		r.RenderColumn(column, x, 0, false)
	}

	if r.Multithreaded {
		r.columnGroup.Done()
	}
}

// Render a frame.
func (r *Renderer) Render() {
	r.RefreshPlayer()
	r.ICacheHits.Store(0)
	r.ICacheMisses.Store(0)

	// Clear buffer, mainly useful for debugging
	/*for i := 0; i < len(r.FrameBuffer); i++ {
		r.FrameBuffer[i][0] = 0
		r.FrameBuffer[i][1] = 0
		r.FrameBuffer[i][2] = 0
		r.FrameBuffer[i][3] = 1
	}*/

	// Frame Tint precalculation
	r.FrameTint = r.Player.FrameTint
	r.FrameTint[0] *= r.FrameTint[3]
	r.FrameTint[1] *= r.FrameTint[3]
	r.FrameTint[2] *= r.FrameTint[3]

	// Make sure we don't have too many debug notices
	for r.DebugNotices.Length() > 30 {
		r.DebugNotices.Pop()
	}
	r.RenderLock.Lock()
	defer r.RenderLock.Unlock()

	r.startingSector = r.PlayerBody.RenderSector()

	if r.Multithreaded {
		blockSize := r.ScreenWidth / r.Blocks
		r.columnGroup.Add(r.Blocks)
		for x := 0; x < r.Blocks; x++ {
			go r.RenderBlock(x, x*blockSize, x*blockSize+blockSize)
		}
		r.columnGroup.Wait()
	} else {
		r.RenderBlock(0, 0, r.ScreenWidth)
	}
	r.RenderHud()

	if r.DB.Frame%4 <= 1 {
		return
	}
	// Invalidate lighting caches
	r.SectorLastRendered.Range(func(eSector concepts.Entity, lastSeen uint64) bool {
		// Cache for a maximum number of frames
		if r.DB.Frame-lastSeen < 120 {
			return true
		}
		if sector := core.SectorFromDb(r.DB, eSector); sector != nil {
			sector.Lightmap.Clear()
		}
		r.SectorLastRendered.Delete(eSector)
		return true
	})
}

func (r *Renderer) ApplyBuffer(buffer []uint8) {
	// TODO: How much faster would a 16-bit integer framebuffer be?
	if r.FrameTint[3] != 0 {
		for fbIndex := 0; fbIndex < r.ScreenWidth*r.ScreenHeight; fbIndex++ {
			screenIndex := fbIndex * 4
			inva := 1.0 - r.FrameTint[3]
			buffer[screenIndex+3] = 0xFF
			buffer[screenIndex+2] = concepts.ByteClamp((r.FrameBuffer[fbIndex][2]*inva + r.FrameTint[2]) * 0xFF)
			buffer[screenIndex+1] = concepts.ByteClamp((r.FrameBuffer[fbIndex][1]*inva + r.FrameTint[1]) * 0xFF)
			buffer[screenIndex+0] = concepts.ByteClamp((r.FrameBuffer[fbIndex][0]*inva + r.FrameTint[0]) * 0xFF)
		}
	} else {
		for fbIndex := 0; fbIndex < r.ScreenWidth*r.ScreenHeight; fbIndex++ {
			screenIndex := fbIndex * 4
			buffer[screenIndex+3] = 0xFF
			buffer[screenIndex+2] = concepts.ByteClamp(r.FrameBuffer[fbIndex][2] * 0xFF)
			buffer[screenIndex+1] = concepts.ByteClamp(r.FrameBuffer[fbIndex][1] * 0xFF)
			buffer[screenIndex+0] = concepts.ByteClamp(r.FrameBuffer[fbIndex][0] * 0xFF)
		}
	}
}

func (r *Renderer) ApplySample(sample *concepts.Vector4, screenIndex int, z float64) {
	if sample[3] <= 0 {
		return
	}
	if sample[3] >= 1 {
		r.FrameBuffer[screenIndex] = *sample
		r.ZBuffer[screenIndex] = z
		return
	}
	inva := 1.0 - sample[3]
	dst := &r.FrameBuffer[screenIndex]
	dst[3] = dst[3]*inva + sample[3]
	if sample[2] < 0 {
		dst[2] *= inva
	} else if sample[2] >= 1 {
		dst[2] = 1
	} else {
		dst[2] = dst[2]*inva + sample[2]
	}
	if sample[1] < 0 {
		dst[1] *= inva
	} else if sample[1] >= 1 {
		dst[1] = 1
	} else {
		dst[1] = dst[1]*inva + sample[1]
	}
	if sample[0] < 0 {
		dst[0] *= inva
	} else if sample[0] >= 1 {
		dst[0] = 1
	} else {
		dst[0] = dst[0]*inva + sample[0]
	}
	if sample[3] > 0.8 {
		r.ZBuffer[screenIndex] = z
	}
}

func (r *Renderer) BitBlt(src *materials.Image, dstx, dsty, w, h int) {
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := x + dstx
			if dx < 0 || dx >= r.ScreenWidth {
				continue
			}
			dy := y + dsty
			if dy < 0 || dy >= r.ScreenHeight {
				continue
			}
			c := src.Sample(float64(x)/float64(w), float64(y)/float64(h), uint32(w), uint32(h))
			r.ApplySample(&c, dx+dy*r.ScreenWidth, -1)
		}
	}
}

func (r *Renderer) Pick(x, y int) []*core.Selectable {
	if x < 0 || y < 0 || x >= r.ScreenWidth || y >= r.ScreenHeight {
		return nil
	}
	// Initialize a column...
	column := &state.Column{
		Config:             r.Config,
		EdgeTop:            0,
		EdgeBottom:         r.ScreenHeight,
		CameraZ:            r.Player.CameraZ,
		PortalColumns:      make([]state.Column, constants.MaxPortals),
		EntitiesByDistance: make([]state.EntityWithDist2, 0, 16),
		Visited:            make([]state.SegmentIntersection, constants.MaxPortals),
	}
	column.LightSampler.MaterialSampler.Config = r.Config

	column.Ray = &state.Ray{Start: *r.PlayerBody.Pos.Render.To2D()}
	column.MaterialSampler = state.MaterialSampler{Config: r.Config, Ray: column.Ray}
	return r.RenderColumn(column, x, y, true)
}
