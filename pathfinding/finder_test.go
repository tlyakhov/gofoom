// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package pathfinding

import (
	"testing"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
)

// Helper to create a sector with manual segments
func createSector(minX, minY, maxX, maxY float64) *core.Sector {
	s := &core.Sector{}
	s.Min = concepts.Vector3{minX, minY, 0}
	s.Max = concepts.Vector3{maxX, maxY, 100}
	s.Bottom.Z.Render = 0
	s.Top.Z.Render = 100

	// Create 4 segments: Bottom, Right, Top, Left (Counter-clockwise)
	// P1(minX, minY) -> P2(maxX, minY) -> P3(maxX, maxY) -> P4(minX, maxY) -> P1

	p1 := concepts.Vector2{minX, minY}
	p2 := concepts.Vector2{maxX, minY}
	p3 := concepts.Vector2{maxX, maxY}
	p4 := concepts.Vector2{minX, maxY}

	points := []concepts.Vector2{p1, p2, p3, p4}
	s.Segments = make([]*core.SectorSegment, 4)

	for i := 0; i < 4; i++ {
		seg := &core.SectorSegment{}
		seg.Sector = s
		seg.P.Render = points[i]
		s.Segments[i] = seg
	}

	// Link segments
	for i := 0; i < 4; i++ {
		curr := s.Segments[i]
		next := s.Segments[(i+1)%4]
		curr.Next = next

		// Set A and B
		curr.A = &curr.P.Render
		curr.B = &next.P.Render

		// Precompute normal and length manually
		// Normal is (dy, -dx) normalized?
		// Segment dir: next - curr
		dx := next.P.Render[0] - curr.P.Render[0]
		dy := next.P.Render[1] - curr.P.Render[1]
		vec := concepts.Vector2{dx, dy}
		length := vec.Length()
		curr.Length = length
		// Normal points INWARDS
		// Rotated 90 degrees CCW? Or CW?
		// P1->P2 (Bottom edge, Y=minY). Points right.
		// Normal should point up (0, 1).
		// (-dy, dx) -> (0, 10). Normalized -> (0, 1). Correct.
		curr.Normal[0] = -dy / length
		curr.Normal[1] = dx / length
	}
	return s
}

func linkSectors(s1, s2 *core.Sector, seg1Idx, seg2Idx int) {
	seg1 := s1.Segments[seg1Idx]
	seg2 := s2.Segments[seg2Idx]

	seg1.AdjacentSector = 1 // Dummy ID
	seg1.AdjacentSegment = seg2

	seg2.AdjacentSector = 2 // Dummy ID
	seg2.AdjacentSegment = seg1
}

func TestSectorForNextPoint(t *testing.T) {
	// Setup: S1 (0,0 to 10,10), S2 (10,0 to 20,10)
	// They share the edge at x=10.
	// S1 segments: 0:Bot, 1:Right(x=10), 2:Top, 3:Left
	// S2 segments: 0:Bot, 1:Right, 2:Top, 3:Left(x=10)

	s1 := createSector(0, 0, 10, 10)
	s2 := createSector(10, 0, 20, 10)

	// Link S1 Right (1) to S2 Left (3)
	linkSectors(s1, s2, 1, 3)

	f := &Finder{
		Start:  &concepts.Vector3{5, 5, 0},
		Step:   1,
		Radius: 0.5,
	}
	f.Request.Ray = &concepts.Ray{}

	// Test 1: Move inside S1
	start := concepts.Vector3{5, 5, 50}
	next := concepts.Vector3{6, 5, 50}

	res := f.sectorForNextPoint(s1, &start, &next)
	if res != s1 {
		t.Errorf("Expected to stay in s1, got %v", res)
	}

	// Test 2: Move from S1 to S2
	// From 9.5, 5 to 10.5, 5
	start = concepts.Vector3{9.5, 5, 50}
	next = concepts.Vector3{10.5, 5, 50}

	res = f.sectorForNextPoint(s1, &start, &next)
	if res != s2 {
		t.Errorf("Expected to move to s2, got %v", res)
	}

	// Test 3: Blocked by wall
	// S1 Top wall at y=10.
	// Move from 5, 9.5 to 5, 10.5
	start = concepts.Vector3{5, 9.5, 50}
	next = concepts.Vector3{5, 10.5, 50}

	res = f.sectorForNextPoint(s1, &start, &next)
	if res != nil {
		t.Errorf("Expected nil (blocked), got %v", res)
	}

	// Test 4: Blocked by radius (close to wall)
	// S1 Left wall at x=0.
	// Move to 0.4, 5. Radius is 0.5. Limit is 0.9.
	// Start at 1.4, 5. Dist = 1.0.
	// Ray hits wall at distance 1.4.
	// Wait, start at 1.4. Target 0.4.
	// Wall at 0.
	// Distance to wall is 1.4.
	// Dist(start, next) = 1.0.
	// Limit = 1.5.
	// Hit at 1.4.
	// 1.4 < 1.5. So hit!
	// Should be blocked.

	start = concepts.Vector3{1.4, 5, 50}
	next = concepts.Vector3{0.4, 5, 50}

	res = f.sectorForNextPoint(s1, &start, &next)
	if res != nil {
		t.Errorf("Expected nil (radius check), got %v", res)
	}
}

// Test KeyToPoint Z preservation
func TestKeyToPointZ(t *testing.T) {
	f := &Finder{
		Start: &concepts.Vector3{10, 10, 50},
		Step:  1,
	}
	p := f.keyToPoint(nodeKey{0, 0})
	if p[2] != 50 {
		t.Errorf("Expected Z=50, got %v", p[2])
	}
}
