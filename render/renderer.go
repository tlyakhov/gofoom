// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package render

import (
	"fmt"
	"math"
	"slices"
	"sync"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/components/materials"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"
	"tlyakhov/gofoom/containers"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"github.com/loov/hrtime"
)

// Renderer holds all state related to a specific camera/map configuration.
type Renderer struct {
	*Config
	Blocks         []block
	blockGroup     *sync.WaitGroup
	startingSector *core.Sector
	textStyle      *TextStyle
	xorSeed        uint64
	tree           *core.Quadtree

	flashOpacity dynamic.DynamicValue[float64]
}

// NewRenderer constructs a new Renderer.
func NewRenderer() *Renderer {
	r := Renderer{
		Config: &Config{
			ScreenWidth:   640,
			ScreenHeight:  360,
			FOV:           constants.FieldOfView,
			Multithreaded: constants.RenderMultiThreaded,
			NumBlocks:     constants.RenderBlocks,
			LightGrid:     constants.LightGrid,
			MaxViewDist:   constants.MaxViewDistance,
		},
		blockGroup: new(sync.WaitGroup),
	}

	r.Initialize()
	return &r
}

func (r *Renderer) Initialize() {
	r.Config.Initialize()

	r.Blocks = make([]block, r.NumBlocks)

	for i := range r.Blocks {
		r.Blocks[i].column.Block = &r.Blocks[i]
		r.Blocks[i].Config = r.Config
		r.Blocks[i].LightLastColHashes = make([]uint64, r.ScreenHeight)
		r.Blocks[i].LightLastColResults = make([]concepts.Vector3, r.ScreenHeight*8)
		r.Blocks[i].LightSampler.tree = ecs.Singleton(core.QuadtreeCID).(*core.Quadtree)
		r.Blocks[i].LightSampler.Visited = make([]*core.Sector, 0, 64)
	}
	r.textStyle = r.NewTextStyle()
	r.xorSeed = concepts.RngXorShift64(uint64(hrtime.Now().Milliseconds()))

	r.flashOpacity.Attach(ecs.Simulation)
	r.flashOpacity.SetAll(0)
	a := r.flashOpacity.NewAnimation()
	r.flashOpacity.Animation = a
	a.Duration = 250
	a.Start = r.flashOpacity.Spawn
	a.End = 1.0
	a.TweeningFunc = dynamic.EaseInOut2
	a.Lifetime = dynamic.AnimationLifetimeBounce
	a.Reverse = false
	a.Active = false
	a.Coordinates = dynamic.AnimationCoordinatesAbsolute
}

func (r *Renderer) WorldToScreen(world *concepts.Vector3) *concepts.Vector2 {
	relative := concepts.Vector2{world[0], world[1]}
	relative[0] -= r.PlayerBody.Pos.Render[0]
	relative[1] -= r.PlayerBody.Pos.Render[1]
	radians := math.Atan2(relative[1], relative[0]) - r.PlayerBody.Angle.Render*concepts.Deg2rad
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

func (r *Renderer) RenderPortal(b *block) {
	// This allocation is ok, does not escape
	portal := &columnPortal{block: b}
	if b.Sector != b.SectorSegment.Sector {
		// We're going into an inner sector
		portal.Adj = b.SectorSegment.Sector
		portal.AdjSegment = b.SectorSegment
	} else if b.SectorSegment.AdjacentSegment != nil {
		// This is an adjacent sector
		portal.Adj = core.GetSector(b.SectorSegment.AdjacentSector)
		portal.AdjSegment = b.SectorSegment.AdjacentSegment
		if b.SectorSegment.AdjacentSegment.PortalTeleports {
			b.teleportRay()
		}
	} else {
		// We are leaving an inner sector
		portal.AdjSegment = b.SectorSegment
		portal.Adj = b.Sector.OuterAt(b.RaySegIntersect.To2D())
	}
	b.LastPortalSegment = portal.AdjSegment

	portal.CalcScreen()
	if portal.AdjSegment != nil {
		if b.Pick {
			wallHiPick(portal)
			wallLowPick(portal)
		} else {
			wallHi(portal)
			wallLow(portal)
		}
	}

	b.EdgeTop = portal.AdjClippedTop
	b.EdgeBottom = portal.AdjClippedBottom
	b.Sector = portal.Adj
	b.LastPortalDistance = b.Distance
	b.Depth++
}

// RenderSegmentColumn draws or picks a single pixel vertical column given a particular
// segment intersection.
func (r *Renderer) RenderSegmentColumn(b *block) {
	b.CalcScreen()

	b.LightSampler.MaterialSampler.Config = r.Config
	b.LightSampler.InputBody = 0
	b.LightSampler.Sector = b.Sector
	b.LightSampler.Segment = nil

	if b.ClippedTop > b.EdgeTop {
		b.LightSampler.Normal = b.Sector.Top.Normal
		if b.Pick {
			planePick(b, &b.Sector.Top)
		} else {
			planes(b, &b.Sector.Top)
		}
	}

	if b.ClippedBottom < b.EdgeBottom {
		b.LightSampler.Normal = b.Sector.Bottom.Normal
		if b.Pick {
			planePick(b, &b.Sector.Bottom)
		} else {
			planes(b, &b.Sector.Bottom)
		}
	}

	if b.ClippedTop >= b.EdgeBottom || b.ClippedBottom <= b.EdgeTop {
		// The segment/portal isn't visible
		return
	}

	b.LightSampler.Segment = b.Segment
	b.Segment.Normal.To3D(&b.LightSampler.Normal)

	hasPortal := b.SectorSegment.AdjacentSector != 0 && b.SectorSegment.AdjacentSegment != nil
	hasPortal = hasPortal || !b.Sector.Outer.Empty()
	if b.Sector != b.SectorSegment.Sector {
		hasPortal = true
		b.LightSampler.Normal.MulSelf(-1)
	}
	switch {
	case hasPortal && !b.SectorSegment.PortalHasMaterial:
		r.RenderPortal(b)
	case b.Pick:
		wallPick(b)
	case hasPortal && b.SectorSegment.PortalHasMaterial:
		saved := b.column
		saved.MaterialSampler.Ray = &saved.Ray
		b.PortalWalls = append(b.PortalWalls, &saved)
		r.RenderPortal(b)
	default:
		r.wall(&b.column)
	}
}

func (r *Renderer) findIntersection(block *block, sector *core.Sector, found bool) bool {
	inner := sector != block.Sector
	for _, sectorSeg := range sector.Segments {
		if sectorSeg == block.LastPortalSegment {
			continue
		}
		// When peeking into inner sectors, ignore their portals. We only care
		// about the ray _entering_ the inner sector.
		if inner && sectorSeg.AdjacentSector != 0 {
			continue
		}

		// Wall is facing away from us
		if inner != (block.Ray.Delta.Dot(&sectorSeg.Normal) > 0) {
			continue
		}

		// Ray intersects?
		u := sectorSeg.Intersect2D(&block.Ray.Start, &block.Ray.End, &block.RaySegTest)
		if u < 0 {
			continue
		}

		// Check if we've already found a closer segment
		dist := block.Ray.DistTo(&block.RaySegTest)
		if (found && dist > block.Distance) ||
			dist <= block.LastPortalDistance {
			continue
		}

		found = true
		block.Segment = &sectorSeg.Segment
		block.SectorSegment = sectorSeg
		block.Distance = dist
		block.RaySegIntersect[0] = block.RaySegTest[0]
		block.RaySegIntersect[1] = block.RaySegTest[1]
		block.segmentIntersection.U = u
	}
	return found
}

// RenderSector intersects a camera ray for a single pixel column with a map sector.
func (r *Renderer) RenderSector(block *block) {
	// Remember the frame # we rendered this sector. This is used when trying to
	// invalidate lighting caches (Sector.Lightmap)
	block.Sector.LastSeenFrame.Store(int64(ecs.Simulation.Frame))

	// Store bodies & internal segments for later
	r.tree.Root.RangeAABB(block.Sector.Min.To2D(), block.Sector.Max.To2D(), func(b *core.Body) bool {
		if b == nil || !b.IsActive() || b.SectorEntity == 0 {
			return true
		}
		block.Bodies.Add(b)
		return true
	})
	for _, iseg := range block.Sector.InternalSegments {
		if iseg == nil || !iseg.IsActive() {
			continue
		}
		block.InternalSegments[iseg] = block.Sector
	}

	// TODO: Fix data race here, since LightmapBias can be read in another goroutine
	if block.Sector.LightmapBias[0] == math.MaxInt64 {
		// Floor is important, needs to truncate towards -Infinity rather than 0
		block.Sector.LightmapBias[2] = int64(math.Floor(block.Sector.Min[2] / r.LightGrid))
		block.Sector.LightmapBias[1] = int64(math.Floor(block.Sector.Min[1] / r.LightGrid))
		block.Sector.LightmapBias[0] = int64(math.Floor(block.Sector.Min[0] / r.LightGrid))
	}

	found := r.findIntersection(block, block.Sector, false)

	for _, inner := range block.Sector.Inner {
		if inner == 0 {
			continue
		}
		found = r.findIntersection(block, core.GetSector(inner), found)
	}

	if found {
		block.IntersectionBottom, block.IntersectionTop = block.Sector.ZAt(block.RaySegIntersect.To2D())
		r.RenderSegmentColumn(block)
	} else {
		dbg := fmt.Sprintf("No intersections for sector %v at depth: %v", block.Sector.Entity, block.Depth)
		r.Player.Notices.Push(dbg)
	}
}

// RenderColumn draws a single pixel column to an 8bit RGBA buffer.
func (r *Renderer) RenderColumn(block *block, x int, y int, pick bool) *PickResult {
	// Reset the z-buffer to maximum viewing distance.
	for i := x; i < r.ScreenHeight*r.ScreenWidth+x; i += r.ScreenWidth {
		r.ZBuffer[i] = r.MaxViewDist
	}

	// Reset the column
	block.LastPortalDistance = 0
	block.LastPortalSegment = nil
	block.LightLastHash = 0
	block.Depth = 0
	block.EdgeTop = 0
	block.EdgeBottom = r.ScreenHeight
	block.Pick = pick
	block.ScreenX = x
	block.ScreenY = y
	block.MaterialSampler.ScreenX = x
	block.MaterialSampler.ScreenY = y
	block.MaterialSampler.Angle = block.Angle
	block.CameraZ = r.Player.CameraZ
	block.ShearZ = r.Player.ShearZ
	block.Ray.Start = *r.PlayerBody.Pos.Render.To2D()
	block.Ray.Set(r.PlayerBody.Angle.Render*concepts.Deg2rad + r.ViewRadians[x])
	block.RayPlane[0] = block.Ray.AngleCos * block.ViewFix[block.ScreenX]
	block.RayPlane[1] = block.Ray.AngleSin * block.ViewFix[block.ScreenX]
	block.PortalWalls = nil

	if r.startingSector != nil {
		block.Sector = r.startingSector
	}
	if block.Sector == nil {
		return nil
	}

	// This used to be recursive, but got expensive for large chains of portals.
	// The iterative approach is faster, but harder to understand as the
	// rendering pipeline manipulates the block/column as it walks the portals.
	for {
		preSector := block.Sector
		r.RenderSector(block)
		if preSector == block.Sector {
			// No more portals
			break
		}
		if block.Depth >= constants.MaxPortals-1 {
			dbg := fmt.Sprintf("Maximum portal depth reached @ %v", block.Sector.Entity)
			r.Player.Notices.Push(dbg)
			break
		}
	}
	if pick {
		return &block.PickResult
	}

	// Draw any walls over portals
	for i := len(block.PortalWalls) - 1; i >= 0; i-- {
		r.wall(block.PortalWalls[i])
	}

	return nil
}

func (r *Renderer) RenderBlock(blockIndex, xStart, xEnd int) {
	// Initialize a block...
	block := &r.Blocks[blockIndex]
	block.LightSampler.xorSeed = r.xorSeed
	block.MaterialSampler = MaterialSampler{Config: r.Config, Ray: &block.Ray}
	ewd2s := make([]*entityWithDistSq, 0, 64)
	block.Bodies = make(containers.Set[*core.Body])
	block.InternalSegments = make(map[*core.InternalSegment]*core.Sector)
	for i := range block.LightLastColHashes {
		block.LightLastColHashes[i] = 0
	}

	/*
		1. Cast rays, render sector walls, ceilings, and floors.
			1.a. Remember which sectors we've seen.
		2. For each sector we've seen:
			2.a. Gather renderable Bodies, collect distances.
			2.b. Gather internal segments, collect distances.
		3. Sort bodies/internal segments by distance.
		4. Render bodies and internal segments in order.
	*/

	for x := xStart; x < xEnd; x++ {
		if x >= r.ScreenWidth {
			break
		}
		r.RenderColumn(block, x, 0, false)
	}

	// Column going through portals affects this
	block.Ray.Start = *r.PlayerBody.Pos.Render.To2D()
	block.CameraZ = r.Player.CameraZ
	block.ShearZ = r.Player.ShearZ

	for b := range block.Bodies {
		vis := materials.GetVisible(b.Entity)
		if vis == nil || !vis.IsActive() {
			continue
		}
		ewd2s = append(ewd2s, &entityWithDistSq{
			Body:    b,
			DistSq:  block.Ray.Start.DistSq(b.Pos.Render.To2D()),
			Visible: vis,
		})
	}
	for iseg, sector := range block.InternalSegments {
		dist := block.Ray.DistTo(iseg.ClosestToPoint(&block.Ray.Start))
		ewd2s = append(ewd2s, &entityWithDistSq{
			InternalSegment: iseg,
			DistSq:          dist * dist,
			Sector:          sector,
		})
	}

	slices.SortFunc(ewd2s, func(a *entityWithDistSq, b *entityWithDistSq) int {
		return int(b.DistSq - a.DistSq)
	})
	// This has a bug when rendering portals: these need to be transformed and
	// clipped through portals appropriately.1
	block.EdgeTop = 0
	block.EdgeBottom = r.ScreenHeight
	block.LightSampler.MaterialSampler.Config = r.Config
	for _, sorted := range ewd2s {
		if sorted.Body != nil {
			r.renderBody(sorted, block, xStart, xEnd)
		} else {
			r.renderInternalSegment(sorted, block, xStart, xEnd)
		}
	}

	if r.Multithreaded {
		r.blockGroup.Done()
	}
}

// Render a frame.
func (r *Renderer) Render() {
	r.tree = ecs.Singleton(core.QuadtreeCID).(*core.Quadtree)
	r.RefreshPlayer()
	if r.PlayerBody == nil {
		return
	}
	LightSamplerCalcs.Store(0)
	LightSamplerLightsTested.Store(0)
	r.xorSeed = concepts.RngXorShift64(r.xorSeed)

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
	for r.Player.Notices.Length() > 30 {
		r.Player.Notices.Pop()
	}
	r.RenderLock.Lock()
	defer r.RenderLock.Unlock()

	r.startingSector = r.PlayerBody.RenderSector()

	if r.Multithreaded {
		blockSize := r.ScreenWidth / r.NumBlocks
		r.blockGroup.Add(r.NumBlocks)
		for x := 0; x < r.NumBlocks; x++ {
			go r.RenderBlock(x, x*blockSize, x*blockSize+blockSize)
		}
		r.blockGroup.Wait()
	} else {
		r.RenderBlock(0, 0, r.ScreenWidth)
	}
	r.RenderHud()
}

func (r *Renderer) ApplyBuffer(buffer []uint8) {
	tm := ecs.Singleton(materials.ToneMapCID).(*materials.ToneMap)

	for i := 0; i < len(r.FrameBuffer); i++ {
		fb := &r.FrameBuffer[i]
		fb[0] = tm.LutLinearToSRGB[concepts.ByteClamp(fb[0]*255)]
		fb[1] = tm.LutLinearToSRGB[concepts.ByteClamp(fb[1]*255)]
		fb[2] = tm.LutLinearToSRGB[concepts.ByteClamp(fb[2]*255)]
	}
	concepts.BlendFrameBuffer(buffer, r.FrameBuffer, &r.FrameTint)
}

func (r *Renderer) BlendSample(sample *concepts.Vector4, screenIndex int, z float64, blendFunc concepts.BlendType) {
	if sample[3] <= 0 {
		return
	}
	concepts.BlendingFuncs[blendFunc](&r.FrameBuffer[screenIndex], sample, 1)
	if sample[3] > 0.8 {
		r.ZBuffer[screenIndex] = z
	}
}

func (r *Renderer) BitBlt(src ecs.Entity, dstx, dsty, w, h int, blendFunc concepts.BlendType) {
	ms := MaterialSampler{
		Config: r.Config,
		ScaleW: uint32(w),
		ScaleH: uint32(h),
	}
	ms.Initialize(src, nil)
	for y := range h {
		for x := range w {
			ms.ScreenX = x + dstx
			if ms.ScreenX < 0 || ms.ScreenX >= r.ScreenWidth {
				continue
			}
			ms.ScreenY = y + dsty
			if ms.ScreenY < 0 || ms.ScreenY >= r.ScreenHeight {
				continue
			}
			ms.NU = float64(x) / float64(w)
			ms.NV = float64(y) / float64(h)
			ms.U = ms.NU
			ms.V = ms.NV
			ms.SampleMaterial(nil)
			screenIndex := ms.ScreenX + ms.ScreenY*r.ScreenWidth
			if blendFunc != concepts.BlendNormal {
				r.BlendSample(&ms.Output, screenIndex, -1, blendFunc)
			} else {
				concepts.BlendColors(&r.FrameBuffer[screenIndex], &ms.Output, 1.0)
			}
		}
	}
}

func (r *Renderer) Pick(x, y int) *PickResult {
	if x < 0 || y < 0 || x >= r.ScreenWidth || y >= r.ScreenHeight {
		return nil
	}
	// Initialize a block..
	block := &block{
		column: column{
			EdgeTop:    0,
			EdgeBottom: r.ScreenHeight,
			CameraZ:    r.Player.CameraZ,
			ShearZ:     r.Player.ShearZ,
		},
		Bodies:           make(containers.Set[*core.Body]),
		InternalSegments: make(map[*core.InternalSegment]*core.Sector),
	}
	block.LightSampler.MaterialSampler.Config = r.Config

	block.Ray = Ray{Start: *r.PlayerBody.Pos.Render.To2D()}
	block.MaterialSampler = MaterialSampler{Config: r.Config, Ray: &block.Ray}
	return r.RenderColumn(block, x, y, true)
}
