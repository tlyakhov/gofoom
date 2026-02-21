package controllers

import (
	"testing"

	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
)

func TestCastCornerTraversal(t *testing.T) {
	// Initialize ECS arena for sectors
	ecs.Initialize()

	// Helper to create a sector
	createSector := func(id ecs.Entity, x, y, size float64) *core.Sector {
		s := ecs.NewAttachedComponent(id, core.SectorCID).(*core.Sector)
		// Manually init essential fields usually handled by Construct
		s.Segments = make([]*core.SectorSegment, 0)
		s.Bodies = make(map[ecs.Entity]*core.Body)
		s.InternalSegments = make(map[ecs.Entity]*core.InternalSegment)
		s.Top.Sector = s
		s.Bottom.Sector = s
		s.Bottom.Normal[2] = 1
		s.Top.Normal[2] = -1

		s.Entity = id
		// Create 4 segments
		// 0: Bottom (y) -> (x, y) to (x+size, y)
		// 1: Right (x+size) -> (x+size, y) to (x+size, y+size)
		// 2: Top (y+size) -> (x+size, y+size) to (x, y+size)
		// 3: Left (x) -> (x, y+size) to (x, y)
		// Note: Winding order usually matters. Let's assume CCW or CW consistent with engine.
		// Engine usually uses CCW for valid area inside? Or logic calculates it.
		// Let's use (0,0)->(10,0)->(10,10)->(0,10)
		s.AddSegment(x, y)           // 0
		s.AddSegment(x+size, y)      // 1
		s.AddSegment(x+size, y+size) // 2
		s.AddSegment(x, y+size)      // 3
		s.Min[2] = 0
		s.Max[2] = 10
		s.Top.Z.SetAll(10)
		s.Bottom.Z.SetAll(0)
		return s
	}

	// Layout:
	//  +----+----+
	//  | TL | TR |  Row 0
	//  +----+----+
	//  | BL | BR |  Row 1
	//  +----+----+
	// (0,0) is top-left in screen coords usually, but let's stick to standard cartesian for simplicity.
	// Let's say:
	// TL: 0,10 to 10,20
	// TR: 10,10 to 20,20
	// BL: 0,0 to 10,10
	// BR: 10,0 to 20,10

	tl := createSector(1, 0, 10, 10)
	tr := createSector(2, 10, 10, 10)
	bl := createSector(3, 0, 0, 10)
	br := createSector(4, 10, 0, 10)

	// Link portals
	// TL Right (Seg 1) <-> TR Left (Seg 3)
	tl.Segments[1].AdjacentSector = tr.Entity
	tl.Segments[1].AdjacentSegment = tr.Segments[3]
	tr.Segments[3].AdjacentSector = tl.Entity
	tr.Segments[3].AdjacentSegment = tl.Segments[1]

	// TL Bottom (Seg 0) <-> BL Top (Seg 2)
	tl.Segments[0].AdjacentSector = bl.Entity
	tl.Segments[0].AdjacentSegment = bl.Segments[2]
	bl.Segments[2].AdjacentSector = tl.Entity
	bl.Segments[2].AdjacentSegment = tl.Segments[0]

	// TR Bottom (Seg 0) <-> BR Top (Seg 2)
	tr.Segments[0].AdjacentSector = br.Entity
	tr.Segments[0].AdjacentSegment = br.Segments[2]
	br.Segments[2].AdjacentSector = tr.Entity
	br.Segments[2].AdjacentSegment = tr.Segments[0]

	// BL Right (Seg 1) <-> BR Left (Seg 3)
	bl.Segments[1].AdjacentSector = br.Entity
	bl.Segments[1].AdjacentSegment = br.Segments[3]
	br.Segments[3].AdjacentSector = bl.Entity
	br.Segments[3].AdjacentSegment = bl.Segments[1]

	// Precompute
	tl.Precompute()
	tr.Precompute()
	bl.Precompute()
	br.Precompute()

	// Place a target body in BR
	target := ecs.NewAttachedComponent(100, core.BodyCID).(*core.Body)
	target.SectorEntity = br.Entity
	target.Pos.Now = concepts.Vector3{15, 5, 5} // Center of BR
	target.Pos.Render = target.Pos.Now
	target.Size.Now = concepts.Vector2{2, 2} // Size is Vector2 (radius, height usually, or width/depth)
	target.Size.Render = target.Size.Now
	br.Bodies[target.Entity] = target

	// Ray from TL center to BR center
	// TL Center: 5, 15
	// BR Center: 15, 5
	// This ray passes exactly through (10, 10), which is the shared corner.
	start := concepts.Vector3{5, 15, 5}
	end := concepts.Vector3{15, 5, 5}

	ray := &concepts.Ray{
		Start: start,
		End:   end,
	}
	ray.AnglesFromStartEnd()

	// Cast
	sel, hit := Cast(ray, tl, 0, false)

	if sel == nil {
		t.Fatalf("Cast failed to hit anything. Expected to hit body in BR.")
	}

	if sel.Body == nil {
		t.Fatalf("Cast hit something, but not a body. Type: %v", sel.Type)
	}

	if sel.Body.Entity != target.Entity {
		t.Fatalf("Cast hit wrong body. Got %v, want %v", sel.Body.Entity, target.Entity)
	}

	// Optional: Check hit point
	dist := hit.Dist(&target.Pos.Render)
	if dist > 2.0 { // Should be close to center
		t.Errorf("Hit point too far from target center: %v", dist)
	}
}
