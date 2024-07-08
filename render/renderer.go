// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"fmt"
	"image"
	"math"
	"slices"
	"sync"

	"github.com/disintegration/imaging"
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
	SectorLastRendered *xsync.MapOf[concepts.Entity, uint64]
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
		Columns:            make([]state.Column, constants.RenderBlocks),
		columnGroup:        new(sync.WaitGroup),
		SectorLastRendered: xsync.NewMapOf[concepts.Entity, uint64](),
	}

	for i := range r.Columns {
		r.Columns[i].Config = r.Config
		r.Columns[i].PortalColumns = make([]state.Column, constants.MaxPortals)
		// Set up 16 slots initially
		r.Columns[i].EntitiesByDistance = make([]state.EntityWithDist2, 0, 16)
	}

	r.Initialize()
	return &r
}

func (r *Renderer) Initialize() {
	r.Config.Initialize()
	for i := range r.Columns {
		r.Columns[i].LightLastColIndices = make([]uint64, r.ScreenHeight)
		r.Columns[i].LightLastColResults = make([]concepts.Vector3, r.ScreenHeight*8)
	}
}

func (r *Renderer) RenderPortal(c *state.Column) {
	if c.Depth >= constants.MaxPortals {
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
		next.RayFloorCeil[0] = next.Ray.AngleCos * c.ViewFix[c.ScreenX]
		next.RayFloorCeil[1] = next.Ray.AngleSin * c.ViewFix[c.ScreenX]
		next.MaterialSampler.Ray = next.Ray
	}

	// This allocation is ok, does not escape
	portal := &state.ColumnPortal{Column: next}
	portal.CalcScreen()
	if portal.AdjSegment != nil {
		if c.Pick {
			WallHiPick(portal)
			WallLowPick(portal)
		} else {
			WallHi(portal)
			WallLow(portal)
		}
	}

	next.Sector = portal.Adj
	next.YStart = portal.AdjClippedTop
	next.YEnd = portal.AdjClippedBottom
	next.LastPortalDistance = c.Distance
	next.Depth++
	r.RenderSector(next)
	c.PickedSelection = next.PickedSelection
}

// RenderSegmentColumn draws or picks a single pixel vertical column given a particular
// segment intersection.
func (r *Renderer) RenderSegmentColumn(c *state.Column) {
	c.CalcScreen()

	c.LightElement.Config = r.Config
	c.LightElement.Type = state.LightElementCeil
	c.LightElement.Normal = c.Sector.CeilNormal
	c.LightElement.Sector = c.Sector
	c.LightElement.Segment = c.Segment

	if c.Pick {
		CeilingPick(c)
	} else {
		Ceiling(c)
	}
	c.LightElement.Type = state.LightElementFloor
	c.LightElement.Normal = c.Sector.FloorNormal

	if c.Pick {
		FloorPick(c)
	} else {
		Floor(c)
	}

	c.LightElement.Type = state.LightElementWall
	c.Segment.Normal.To3D(&c.LightElement.Normal)

	hasPortal := c.SectorSegment.AdjacentSector != 0
	if c.Pick {
		if !hasPortal || c.SectorSegment.PortalHasMaterial {
			WallMidPick(c)
			return
		}
		r.RenderPortal(c)
	} else {
		if hasPortal {
			r.RenderPortal(c)
		}
		if !hasPortal || c.SectorSegment.PortalHasMaterial {
			WallMid(c, false)
		}
	}

}

// RenderSector intersects a camera ray for a single pixel column with a map sector.
func (r *Renderer) RenderSector(c *state.Column) {
	// This is for invalidating lighting caches (Sector.Lightmap)
	r.SectorLastRendered.Store(c.Sector.Entity, uint64(r.Frame))
	c.Distance = constants.MaxViewDistance
	c.SectorSegment = nil
	c.Segment = nil
	for _, sectorSeg := range c.Sector.Segments {
		if !c.IntersectSegment(&sectorSeg.Segment, true, false) {
			continue
		}
		c.SectorSegment = sectorSeg
		c.BottomZ, c.TopZ = c.Sector.SlopedZRender(c.RaySegIntersect.To2D())
	}

	if c.Segment != nil {
		r.RenderSegmentColumn(c)
	} else {
		dbg := fmt.Sprintf("No intersections for sector %v at depth: %v", c.Sector.Entity, c.Depth)
		r.DebugNotices.Push(dbg)
	}

	// Clear slice without reallocating memory
	c.EntitiesByDistance = c.EntitiesByDistance[:0]
	for entity, b := range c.Sector.Bodies {
		c.EntitiesByDistance = append(c.EntitiesByDistance, state.EntityWithDist2{
			Entity:    entity,
			Dist2:     c.Ray.Start.Dist2(b.Pos.Render.To2D()),
			IsSegment: false,
		})
	}
	for entity, s := range c.Sector.InternalSegments {
		// TODO: we do this again later. Should we optimize this?
		if s == nil || !s.IsActive() || !c.IntersectSegment(&s.Segment, false, s.TwoSided) {
			continue
		}
		c.EntitiesByDistance = append(c.EntitiesByDistance, state.EntityWithDist2{
			Entity:    entity,
			Dist2:     c.Distance * c.Distance,
			IsSegment: true,
		})
	}

	slices.SortFunc(c.EntitiesByDistance, func(a state.EntityWithDist2, b state.EntityWithDist2) int {
		return int(b.Dist2 - a.Dist2)
	})
	c.SectorSegment = nil
	for _, sorted := range c.EntitiesByDistance {
		if sorted.Entity == 0 {
			continue
		}
		if !sorted.IsSegment {
			r.RenderBody(sorted.Entity, c)
			continue
		}

		s := core.InternalSegmentFromDb(r.DB, sorted.Entity)
		c.IntersectSegment(&s.Segment, false, s.TwoSided)
		c.TopZ = s.Top
		c.BottomZ = s.Bottom
		c.CalcScreen()

		if c.Pick && c.ScreenY >= c.ClippedStart && c.ScreenY <= c.ClippedEnd {
			c.PickedSelection = append(c.PickedSelection, core.SelectableFromInternalSegment(s))
			return
		}
		c.LightElement.Config = r.Config
		c.LightElement.Sector = c.Sector
		c.LightElement.Segment = &s.Segment
		c.LightElement.Type = state.LightElementWall
		s.Segment.Normal.To3D(&c.LightElement.Normal)
		WallMid(c, true)
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
	column.YStart = 0
	column.YEnd = r.ScreenHeight
	column.Pick = pick
	column.ScreenX = x
	column.ScreenY = y
	column.MaterialSampler.ScreenX = x
	column.MaterialSampler.ScreenY = y
	column.MaterialSampler.Angle = column.Angle
	column.Ray.Set(r.PlayerBody.Angle.Render*concepts.Deg2rad + r.ViewRadians[x])
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

func (r *Renderer) RenderBlock(buffer []uint8, columnIndex, xStart, xEnd int) {
	bob := math.Sin(r.Player.Bob)
	// Initialize a column...
	column := &r.Columns[columnIndex]
	column.CameraZ = r.PlayerBody.Pos.Render[2] + r.PlayerBody.Size.Render[1]*0.5 + bob
	column.Ray = &state.Ray{Start: *r.PlayerBody.Pos.Render.To2D()}
	column.MaterialSampler = state.MaterialSampler{Config: r.Config, Ray: column.Ray}
	for i := range column.LightLastColIndices {
		column.LightLastColIndices[i] = 0
	}

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
	r.Frame++
	r.Counter = 0

	if constants.RenderMultiThreaded {
		blockSize := r.ScreenWidth / constants.RenderBlocks
		r.columnGroup.Add(constants.RenderBlocks)
		for x := 0; x < constants.RenderBlocks; x++ {
			go r.RenderBlock(buffer, x, x*blockSize, x*blockSize+blockSize)
		}
		r.columnGroup.Wait()
	} else {
		r.RenderBlock(buffer, 0, 0, r.ScreenWidth)
	}
	r.RenderHud(buffer)
	// Invalidate lighting caches
	r.SectorLastRendered.Range(func(eSector concepts.Entity, lastSeen uint64) bool {
		// Cache for a maximum number of frames
		if r.Frame-lastSeen < 120 {
			return true
		}
		if sector := r.DB.Component(eSector, core.SectorComponentIndex).(*core.Sector); sector != nil {
			sector.Lightmap.Clear()
		}
		r.SectorLastRendered.Delete(eSector)
		return true
	})
}

func (r *Renderer) ImgBlt(dst []uint8, src *image.NRGBA, dstx, dsty int) {
	w := src.Rect.Dx()
	h := src.Rect.Dy()
	idst := (dstx + dsty*r.ScreenWidth) * 4
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			isrc := x*4 + y*src.Stride
			a := int(src.Pix[isrc+3])
			da := 255 - a
			dst[idst+0] = uint8((int(dst[idst+0]) * da / 255) + int(src.Pix[isrc+0])*a/255)
			dst[idst+1] = uint8((int(dst[idst+1]) * da / 255) + int(src.Pix[isrc+1])*a/255)
			dst[idst+2] = uint8((int(dst[idst+2]) * da / 255) + int(src.Pix[isrc+2])*a/255)
			idst += 4
		}
		idst = (dstx + (dsty+y)*r.ScreenWidth) * 4
	}
}

func (r *Renderer) RenderHud(buffer []uint8) {
	if r.Player == nil {
		return
	}
	for _, item := range r.Player.Inventory {
		img := materials.ImageFromDb(r.DB, item.Image)
		if img == nil {
			return
		}
		rimg := imaging.Resize(img.Image, 32, 0, imaging.CatmullRom)
		r.ImgBlt(buffer, rimg, 10, r.ScreenHeight-42)
	}
}

func (r *Renderer) Pick(x, y int) []*core.Selectable {
	if x < 0 || y < 0 || x >= r.ScreenWidth || y >= r.ScreenHeight {
		return nil
	}
	bob := math.Sin(r.Player.Bob) * 2
	// Initialize a column...
	column := &state.Column{
		Config:             r.Config,
		YStart:             0,
		YEnd:               r.ScreenHeight,
		CameraZ:            r.PlayerBody.Pos.Render[2] + r.PlayerBody.Size.Render[1]*0.5 + bob,
		PortalColumns:      make([]state.Column, constants.MaxPortals),
		EntitiesByDistance: make([]state.EntityWithDist2, 0, 16),
	}
	column.LightElement.Config = r.Config

	column.Ray = &state.Ray{Start: *r.PlayerBody.Pos.Render.To2D()}
	column.MaterialSampler = state.MaterialSampler{Config: r.Config, Ray: column.Ray}
	return r.RenderColumn(column, x, y, true)
}
