package core

import (
	"fmt"
	"sort"

	"tlyakhov/gofoom/concepts"

	"github.com/rs/xid"
)

// The algorithm here is from:
// https://geidav.wordpress.com/2015/03/21/splitting-an-arbitrary-polygon-by-a-line/

type splitSide int

const (
	sideLeft  splitSide = -1
	sideOn              = 0
	sideRight           = 1
)

type SectorSplitter struct {
	SplitSector          []*splitEdge
	EdgesOnLine          []*splitEdge
	Splitter1, Splitter2 concepts.Vector2
	Sector               AbstractSector
	Result               []AbstractSector
}

type splitEdge struct {
	*SectorSplitter
	Source           *Segment
	Start            concepts.Vector2
	Side             splitSide
	Next             *splitEdge
	Prev             *splitEdge
	DistOnLine       float64
	SrcEdge, DstEdge bool
	Visited          bool
}

func (se *splitEdge) String() string {
	p := concepts.Vector2{}
	n := concepts.Vector2{}
	if se.Prev != nil {
		p = se.Prev.Start
	}
	if se.Next != nil {
		n = se.Next.Start
	}
	return fmt.Sprintf("Split Edge: <%v>, dist: %v, side: %v, src/dest: %v/%v, prev: <%v>, next: <%v>", se.Start.String(), se.DistOnLine, se.Side, se.SrcEdge, se.DstEdge, p, n)
}

type splitEdgeByStart []*splitEdge

func (edges splitEdgeByStart) Len() int      { return len(edges) }
func (edges splitEdgeByStart) Swap(i, j int) { edges[i], edges[j] = edges[j], edges[i] }
func (edges splitEdgeByStart) Less(i, j int) bool {
	return edges[i].signedDist(&edges[i].Splitter1, &edges[i].Splitter2, &edges[i].Start) < edges[j].signedDist(&edges[j].Splitter1, &edges[j].Splitter2, &edges[j].Start)
}

func (a *SectorSplitter) whichSide(l1, l2, p *concepts.Vector2) splitSide {
	ld := l2.Sub(l1)
	pd := p.Sub(l1)
	d := pd[0]*ld[1] - pd[1]*ld[0]

	if d > 0.000001 {
		return sideRight
	} else if d < -0.000001 {
		return sideLeft
	}
	return sideOn
}

func (a *SectorSplitter) signedDist(l1, l2, p *concepts.Vector2) float64 {
	return l2.Sub(l1).Dot(p.Sub(l1))
}

func (a *SectorSplitter) Do() {
	a.splitEdges()
	if len(a.EdgesOnLine) > 0 {
		a.sortEdges()
		a.split()
		a.collect()
	} else {
		a.Result = nil
	}
}

func (a *SectorSplitter) splitEdges() {
	// fmt.Printf("Splitting %v\n", a.Sector.GetBase().ID)

	a.SplitSector = []*splitEdge{}
	a.EdgesOnLine = []*splitEdge{}

	// We need to iterate counter-clockwise for the splitting to work, so let's use the sector winding to figure that out.
	start := 0
	end := len(a.Sector.Physical().Segments) - 1
	dir := int(a.Sector.Physical().Winding)
	if a.Sector.Physical().Winding < 0 {
		end, start = start, end
	}

	for i := start; i != end+dir; i += dir {
		j := i + dir
		if j < 0 {
			j += len(a.Sector.Physical().Segments)
		} else if j >= len(a.Sector.Physical().Segments) {
			j -= len(a.Sector.Physical().Segments)
		}
		segment := a.Sector.Physical().Segments[i]
		next := a.Sector.Physical().Segments[j]
		edgeStartSide := a.whichSide(&a.Splitter1, &a.Splitter2, &segment.P)
		edgeEndSide := a.whichSide(&a.Splitter1, &a.Splitter2, &next.P)
		se := &splitEdge{SectorSplitter: a, Source: segment, Start: segment.P, Side: edgeStartSide}
		a.SplitSector = append(a.SplitSector, se)
		// fmt.Printf("Added %v to SplitSector...\n", se.String())

		if edgeStartSide == sideOn {
			a.EdgesOnLine = append(a.EdgesOnLine, se)
			// fmt.Printf("Edge on line!\n")
		} else if edgeStartSide != edgeEndSide && edgeEndSide != sideOn {
			isect, ok := concepts.Intersect(&segment.P, &next.P, &a.Splitter1, &a.Splitter2)
			// fmt.Printf("Edge intersects at %v\n", isect.String())
			if !ok {
				// The splitter line is not fully bisecting the sector. Ignore, and continue.
				// fmt.Println("Splitter not bisecting sector.")
				continue
			}
			se := &splitEdge{SectorSplitter: a, Source: segment, Start: *isect, Side: sideOn}
			a.SplitSector = append(a.SplitSector, se)
			a.EdgesOnLine = append(a.EdgesOnLine, se)
		} else {
			// fmt.Printf("Edge doesn't intersect split line (or end point is on the line).\n")
		}
	}

	// fmt.Println("Final constructed splitter:")
	// Connect doubly linked list
	for i, edge := range a.SplitSector {
		next := a.SplitSector[(i+1)%len(a.SplitSector)]
		edge.Next = next
		next.Prev = edge
		// fmt.Printf("%v\n", edge.String())
	}
}

func (a *SectorSplitter) sortEdges() {
	// Sort edges by start position relative to the start position of the split line
	sort.Sort(splitEdgeByStart(a.EdgesOnLine))

	// fmt.Println("Sorted edges:")
	// Compute the distance of each edge to the first one.
	for _, edge := range a.EdgesOnLine {
		edge.DistOnLine = edge.Start.Dist(&a.EdgesOnLine[0].Start)
		// fmt.Printf("%v\n", edge.String())
	}
}

func (a *SectorSplitter) split() {
	var useSrc *splitEdge

	for i := 0; i < len(a.EdgesOnLine); i++ {
		srcEdge := useSrc
		useSrc = nil

		for ; srcEdge == nil && i < len(a.EdgesOnLine); i++ {
			edge := a.EdgesOnLine[i]
			if edge.Side != sideOn {
				panic("Split sector error: edge in EdgesOnLine not SideOn.")
			}

			if (edge.Prev.Side == sideLeft && edge.Next.Side == sideRight) ||
				(edge.Prev.Side == sideLeft && edge.Next.Side == sideOn && edge.Next.DistOnLine < edge.DistOnLine) ||
				(edge.Prev.Side == sideOn && edge.Next.Side == sideRight && edge.Prev.DistOnLine < edge.DistOnLine) {
				srcEdge = edge
				srcEdge.SrcEdge = true
			}
		}

		var dstEdge *splitEdge

		for dstEdge == nil && i < len(a.EdgesOnLine) {
			edge := a.EdgesOnLine[i]
			if edge.Side != sideOn {
				panic("Split sector error: edge in EdgesOnLine not SideOn.")
			}

			if (edge.Prev.Side == sideRight && edge.Next.Side == sideLeft) ||
				(edge.Prev.Side == sideOn && edge.Next.Side == sideLeft) ||
				(edge.Prev.Side == sideRight && edge.Next.Side == sideOn) ||
				(edge.Prev.Side == sideRight && edge.Next.Side == sideRight) ||
				(edge.Prev.Side == sideLeft && edge.Next.Side == sideLeft) {

				dstEdge = edge
				dstEdge.DstEdge = true
			} else {
				i++
			}
		}

		// Bridge source and destination
		if srcEdge != nil && dstEdge != nil {
			a.createBridge(srcEdge, dstEdge)
			a.verifyCycles()

			// Is it a config in which a vertex needs to be reused as a source vertex?
			if srcEdge.Prev.Prev.Side == sideLeft {
				useSrc = srcEdge.Prev
				useSrc.SrcEdge = true
			} else if dstEdge.Next.Side == sideRight {
				useSrc = dstEdge
				useSrc.SrcEdge = true
			}
		}
	}
}

func (a *SectorSplitter) createBridge(srcEdge, dstEdge *splitEdge) {
	src2 := *srcEdge
	dst2 := *dstEdge
	a.SplitSector = append(a.SplitSector, &src2, &dst2)
	src2.Next = dstEdge
	src2.Prev = srcEdge.Prev
	dst2.Next = srcEdge
	dst2.Prev = dstEdge.Prev
	srcEdge.Prev.Next = &src2
	srcEdge.Prev = &dst2
	dstEdge.Prev.Next = &dst2
	dstEdge.Prev = &src2
}

func (a *SectorSplitter) verifyCycles() {
	// fmt.Println("Verifying cycles...")
	for _, edge := range a.SplitSector {
		visitor := edge
		count := 0

		// fmt.Printf("Starting cycle test at %v\n", edge.String())
		for {
			if count > len(a.SplitSector) {
				panic("Split Sector error: verify cycles failed.")
			}
			// fmt.Printf("%v\n", visitor.String())
			visitor = visitor.Next
			count++
			if visitor == edge {
				break
			}
		}
	}
}

func (a *SectorSplitter) collect() {
	a.Result = []AbstractSector{}

	for _, edge := range a.SplitSector {
		if edge.Visited {
			continue
		}

		// Clone the original sector using serialization.
		added := concepts.MapPolyStruct(a.Sector.Physical().Map, a.Sector.Serialize()).(AbstractSector)
		added.GetBase().ID = xid.New().String()
		a.Result = append(a.Result, added)
		phys := added.Physical()
		// Don't clone the mobs.
		phys.Mobs = make(map[string]AbstractMob)
		// Clear segments
		phys.Segments = []*Segment{}

		visitor := edge
		for {
			visitor.Visited = true
			addedSegment := &Segment{}
			addedSegment.SetParent(added)
			addedSegment.Construct(visitor.Source.Serialize())
			addedSegment.GetBase().ID = xid.New().String()
			addedSegment.P = visitor.Start
			addedSegment.AdjacentSegment = nil
			addedSegment.AdjacentSector = nil
			if len(phys.Segments) > 0 {
				addedSegment.Next = phys.Segments[0]
				addedSegment.Prev = phys.Segments[len(phys.Segments)-1]
				addedSegment.Next.Prev = addedSegment
				addedSegment.Prev.Next = addedSegment
			}
			phys.Segments = append(phys.Segments, addedSegment)
			if visitor.Source.AdjacentSegment != nil {
				visitor.Source.AdjacentSegment.AdjacentSector = nil
				visitor.Source.AdjacentSegment.AdjacentSegment = nil
			}
			visitor = visitor.Next
			if visitor == edge {
				break
			}
		}
	}
	// Only one sector means we didn't split anything
	if len(a.Result) == 1 && len(a.Result[0].Physical().Segments) == len(a.Sector.Physical().Segments) {
		a.Result = nil
		return
	}

	for id, e := range a.Sector.Physical().Mobs {
		e.Physical().Sector = nil
		for _, added := range a.Result {
			phys := added.Physical()
			phys.Recalculate()
			if phys.IsPointInside2D(&concepts.Vector2{e.Physical().Pos.Original[0], e.Physical().Pos.Original[1]}) {
				phys.Mobs[id] = e
				e.SetParent(added)
			}
		}
	}
}
