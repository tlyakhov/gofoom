package core

import (
	"sort"

	"github.com/rs/xid"
	"github.com/tlyakhov/gofoom/concepts"
)

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

type splitEdgeByStart []*splitEdge

func (edges splitEdgeByStart) Len() int      { return len(edges) }
func (edges splitEdgeByStart) Swap(i, j int) { edges[i], edges[j] = edges[j], edges[i] }
func (edges splitEdgeByStart) Less(i, j int) bool {
	return edges[i].signedDist(edges[i].Splitter1, edges[i].Splitter2, edges[i].Start) < edges[i].signedDist(edges[i].Splitter1, edges[i].Splitter2, edges[j].Start)
}

func (a *SectorSplitter) whichSide(l1, l2, p concepts.Vector2) splitSide {
	ld := l2.Sub(l1)
	pd := p.Sub(l1)
	d := pd.X*ld.Y - pd.Y*ld.X

	if d > 0.000001 {
		return sideRight
	} else if d < -0.000001 {
		return sideLeft
	}
	return sideOn
}

func (a *SectorSplitter) signedDist(l1, l2, p concepts.Vector2) float64 {
	return l2.Sub(l1).Dot(p.Sub(l1))
}

func (a *SectorSplitter) Do() {
	a.splitEdges()
	a.sortEdges()
	a.split()
	a.collect()
}

func (a *SectorSplitter) splitEdges() {
	a.SplitSector = []*splitEdge{}
	a.EdgesOnLine = []*splitEdge{}

	for _, segment := range a.Sector.Physical().Segments {
		edgeStartSide := a.whichSide(a.Splitter1, a.Splitter2, segment.A)
		edgeEndSide := a.whichSide(a.Splitter1, a.Splitter2, segment.B)
		se := &splitEdge{SectorSplitter: a, Source: segment, Start: segment.A, Side: edgeStartSide}
		a.SplitSector = append(a.SplitSector, se)

		if edgeStartSide == sideOn {
			a.EdgesOnLine = append(a.EdgesOnLine, se)
		} else if edgeStartSide != edgeEndSide && edgeEndSide != sideOn {
			isect, ok := segment.Intersect2D(a.Splitter1, a.Splitter2)
			if !ok {
				panic("Split Sector error: no intersection despite side difference!")
			}
			se := &splitEdge{SectorSplitter: a, Source: segment, Start: isect, Side: sideOn}
			a.SplitSector = append(a.SplitSector, se)
			a.EdgesOnLine = append(a.EdgesOnLine, se)
		}
	}

	// Connect doubly linked list
	for i, edge := range a.SplitSector {
		edge.Next = a.SplitSector[(i+1)%len(a.SplitSector)]
		a.SplitSector[(i+1)%len(a.SplitSector)].Prev = edge
	}
}

func (a *SectorSplitter) sortEdges() {
	// Sort edges by start position relative to the start position of the split line
	sort.Sort(splitEdgeByStart(a.EdgesOnLine))

	// Compute the distance of each edge to the first one.
	for _, edge := range a.EdgesOnLine {
		edge.DistOnLine = edge.Start.Dist(a.EdgesOnLine[0].Start)
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
	for _, edge := range a.SplitSector {
		visitor := edge
		count := 0

		for {
			if count >= len(a.SplitSector) {
				panic("Split Sector error: verify cycles failed.")
			}
			visitor = edge.Next
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
		// Don't clone the entities.
		phys.Entities = make(map[string]AbstractEntity)
		// Clear segments
		phys.Segments = []*Segment{}

		visitor := edge
		for {
			visitor.Visited = true
			addedSegment := &Segment{}
			addedSegment.SetParent(added)
			addedSegment.Deserialize(visitor.Source.Serialize())
			addedSegment.GetBase().ID = xid.New().String()
			addedSegment.A = visitor.Start
			addedSegment.AdjacentSegment = nil
			addedSegment.AdjacentSector = nil
			phys.Segments = append(phys.Segments, addedSegment)
			visitor = visitor.Next
			if visitor == edge {
				break
			}
		}
	}
}
