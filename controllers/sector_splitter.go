// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package controllers

import (
	"fmt"
	"sort"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

// The algorithm here is from:
// https://geidav.wordpress.com/2015/03/21/splitting-an-arbitrary-polygon-by-a-line/

type splitSide int

const (
	sideLeft  splitSide = -1
	sideOn    splitSide = 0
	sideRight splitSide = 1
)

type SectorSplitter struct {
	SplitSector          []*splitEdge
	EdgesOnLine          []*splitEdge
	Splitter1, Splitter2 concepts.Vector2
	Sector               *core.Sector
	Result               [][]ecs.Attachable
}

type splitEdge struct {
	*SectorSplitter
	Source           *core.SectorSegment
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
	// fmt.Printf("Splitting %v\n", a.a.Sector.GetBase().Name)

	a.SplitSector = []*splitEdge{}
	a.EdgesOnLine = []*splitEdge{}

	// We need to iterate counter-clockwise for the splitting to work, so let's use the a.Sector winding to figure that out.
	start := 0
	end := len(a.Sector.Segments) - 1
	dir := int(a.Sector.Winding)
	if a.Sector.Winding < 0 {
		end, start = start, end
	}

	for i := start; i != end+dir; i += dir {
		j := i + dir
		if j < 0 {
			j += len(a.Sector.Segments)
		} else if j >= len(a.Sector.Segments) {
			j -= len(a.Sector.Segments)
		}
		segment := a.Sector.Segments[i]
		next := a.Sector.Segments[j]
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
				// The splitter line is not fully bisecting the a.Sector. Ignore, and continue.
				// fmt.Println("Splitter not bisecting a.Sector.")
				continue
			}
			se := &splitEdge{SectorSplitter: a, Source: segment, Start: *isect, Side: sideOn}
			a.SplitSector = append(a.SplitSector, se)
			a.EdgesOnLine = append(a.EdgesOnLine, se)
		} /*else {
			fmt.Printf("Edge doesn't intersect split line (or end point is on the line).\n")
		}*/
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
				panic("Split a.Sector error: edge in EdgesOnLine not SideOn.")
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
				panic("Split a.Sector error: edge in EdgesOnLine not SideOn.")
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
				panic("Split a.Sector error: verify cycles failed.")
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
	a.Result = make([][]ecs.Attachable, 0)
	newSectorCount := 0
	for _, edge := range a.SplitSector {
		if edge.Visited {
			continue
		}

		// Clone all the original components using serialization.
		newSectorCount++
		db := a.Sector.ECS
		newEntity := db.NewEntity()
		clonedComponents := make([]ecs.Attachable, len(db.AllComponents(a.Sector.Entity)))
		a.Result = append(a.Result, clonedComponents)
		for componentIndex, origComponent := range db.AllComponents(a.Sector.Entity) {
			if origComponent == nil {
				continue
			}
			addedComponent := db.NewAttachedComponent(newEntity, componentIndex)
			// This will override the entity field with the original component's
			// entity, so we'll fix it up afterwards.
			addedComponent.Construct(origComponent.Serialize())
			addedComponent.SetEntity(newEntity)

			clonedComponents[componentIndex] = addedComponent
			switch target := addedComponent.(type) {
			case *ecs.Named:
				target.Name = fmt.Sprintf("Split %v (%v)", target.Name, newSectorCount)
			case *core.Sector:
				// Don't clone the bodies.
				target.Bodies = make(map[ecs.Entity]*core.Body)
				// Clear segments
				target.Segments = []*core.SectorSegment{}

				visitor := edge
				for {
					visitor.Visited = true
					addedSegment := &core.SectorSegment{}
					addedSegment.Sector = target
					addedSegment.Construct(target.ECS, visitor.Source.Serialize())
					addedSegment.P = visitor.Start
					target.Segments = append(target.Segments, addedSegment)
					if visitor.Source.AdjacentSegment != nil {
						visitor.Source.AdjacentSegment.AdjacentSector = target.Entity
						visitor.Source.AdjacentSegment.AdjacentSegment = addedSegment
					}
					visitor = visitor.Next
					if visitor == edge {
						break
					}
				}
				target.Recalculate()
			}
		}
	}
	// Only one a.Sector means we didn't split anything
	if len(a.Result) == 1 {
		added := a.Result[0][core.SectorComponentIndex].(*core.Sector)
		if len(added.Segments) == len(a.Sector.Segments) {
			a.Result = nil
			return
		}
	}

	col := ecs.ColumnFor[core.Body](a.Sector.ECS, core.BodyComponentIndex)
	for i := range col.Length {
		body := col.Value(i)
		if body.SectorEntity != a.Sector.Entity {
			continue
		}
		for _, components := range a.Result {
			if components[core.SectorComponentIndex] == nil {
				continue
			}
			if added, ok := components[core.SectorComponentIndex].(*core.Sector); ok &&
				added.IsPointInside2D(body.Pos.Original.To2D()) {
				body.SectorEntity = added.Entity
				added.Bodies[body.Entity] = body
			}
		}
	}
}
