// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"fmt"
	"testing"
)

// Mock component that allows setting the ComponentID
type MockComponent struct {
	Attached
	id ComponentID
}

func (m *MockComponent) ComponentID() ComponentID {
	return m.id
}

func (m *MockComponent) String() string {
	return fmt.Sprintf("MockComponent(%d)", m.id)
}

// Ensure MockComponent satisfies the Component interface
var _ Component = &MockComponent{}

func TestComponentTable_BasicSetGet(t *testing.T) {
	var table ComponentTable

	c1 := &MockComponent{id: 1}
	c2 := &MockComponent{id: 2}

	table.Set(c1)
	table.Set(c2)

	if got := table.Get(1); got != c1 {
		t.Errorf("Get(1) = %v; want %v", got, c1)
	}
	if got := table.Get(2); got != c2 {
		t.Errorf("Get(2) = %v; want %v", got, c2)
	}

	// Update existing
	c1New := &MockComponent{id: 1}
	table.Set(c1New)
	if got := table.Get(1); got != c1New {
		t.Errorf("Get(1) after update = %v; want %v", got, c1New)
	}
}

func TestComponentTable_Delete(t *testing.T) {
	var table ComponentTable
	c1 := &MockComponent{id: 1}

	table.Set(c1)
	if got := table.Get(1); got != c1 {
		t.Fatalf("Setup failed: Get(1) = %v; want %v", got, c1)
	}

	table.Delete(1)
	if got := table.Get(1); got != nil {
		t.Errorf("Get(1) after delete = %v; want nil", got)
	}
}

func TestComponentTable_Collisions(t *testing.T) {
	var table ComponentTable
	// ComponentTableGrowthRate is 8.
	// We want collisions. id % size. Initial size is 8.
	// 1 % 8 = 1
	// 9 % 8 = 1
	// 17 % 8 = 1

	c1 := &MockComponent{id: 1}
	c9 := &MockComponent{id: 9}
	c17 := &MockComponent{id: 17}

	table.Set(c1)
	table.Set(c9)
	table.Set(c17)

	if got := table.Get(1); got != c1 {
		t.Errorf("Get(1) = %v; want %v", got, c1)
	}
	if got := table.Get(9); got != c9 {
		t.Errorf("Get(9) = %v; want %v", got, c9)
	}
	if got := table.Get(17); got != c17 {
		t.Errorf("Get(17) = %v; want %v", got, c17)
	}
}

func TestComponentTable_DeleteCollision(t *testing.T) {
	var table ComponentTable
	// Setup collision chain: 1, 9, 17 (all hash to 1 mod 8)
	c1 := &MockComponent{id: 1}
	c9 := &MockComponent{id: 9}
	c17 := &MockComponent{id: 17}

	table.Set(c1)
	table.Set(c9)
	table.Set(c17)

	// Delete the middle one
	table.Delete(9)

	if got := table.Get(1); got != c1 {
		t.Errorf("Get(1) = %v; want %v", got, c1)
	}
	if got := table.Get(9); got != nil {
		t.Errorf("Get(9) = %v; want nil", got)
	}
	if got := table.Get(17); got != c17 {
		t.Errorf("Get(17) = %v; want %v", got, c17)
	}

	// Delete the head
	table.Delete(1)
	if got := table.Get(1); got != nil {
		t.Errorf("Get(1) = %v; want nil", got)
	}
	if got := table.Get(17); got != c17 {
		t.Errorf("Get(17) = %v; want %v", got, c17)
	}
}

func TestComponentTable_Growth(t *testing.T) {
	var table ComponentTable
	// Initial size 8. Add more than 8 elements.
	count := 20
	for i := 1; i <= count; i++ {
		table.Set(&MockComponent{id: ComponentID(i)})
	}

	for i := 1; i <= count; i++ {
		got := table.Get(ComponentID(i))
		if got == nil {
			t.Errorf("Get(%d) = nil; want component", i)
		} else if got.ComponentID() != ComponentID(i) {
			t.Errorf("Get(%d).ID = %d; want %d", i, got.ComponentID(), i)
		}
	}
}

func TestComponentTable_NonExistent(t *testing.T) {
	var table ComponentTable

	// Empty table
	if got := table.Get(1); got != nil {
		t.Errorf("Empty table Get(1) = %v; want nil", got)
	}
	table.Delete(1) // Should not panic

	c1 := &MockComponent{id: 1}
	table.Set(c1)

	if got := table.Get(2); got != nil {
		t.Errorf("Get(2) = %v; want nil", got)
	}
	table.Delete(2) // Should not panic and not delete 1

	if got := table.Get(1); got != c1 {
		t.Errorf("Get(1) = %v; want %v", got, c1)
	}
}

func TestComponentTable_DeleteRehashComplex(t *testing.T) {
	// This test aims to create a scenario where deleting an element
	// requires moving another element that is "far" away due to probing.
	var table ComponentTable

	// Size 8.
	// Slot 1: id=1
	// Slot 2: id=2
	// Slot 3: id=9 (hash 1, probed to 3 because 2 is taken? No, 2 hashes to 2)

	// Let's construct:
	// id=1 (hash 1) -> slot 1
	// id=9 (hash 1) -> slot 2
	// id=17 (hash 1) -> slot 3

	c1 := &MockComponent{id: 1}
	c9 := &MockComponent{id: 9}
	c17 := &MockComponent{id: 17}

	table.Set(c1)
	table.Set(c9)
	table.Set(c17)

	// Table should look like:
	// [0]: nil
	// [1]: c1
	// [2]: c9
	// [3]: c17
	// ...

	// Delete c1. c9 should move to slot 1. c17 should move to slot 2.
	table.Delete(1)

	if table.Get(9) != c9 {
		t.Error("Get(9) failed after deleting 1")
	}
	if table.Get(17) != c17 {
		t.Error("Get(17) failed after deleting 1")
	}
}

func TestComponentTable_FullWrapAround(t *testing.T) {
	var table ComponentTable
	// Force wrap around probing.
	// Size 8.
	// Fill 6, 7.
	// Add something that hashes to 6 or 7, should wrap to 0.

	c6 := &MockComponent{id: 6}
	c7 := &MockComponent{id: 7}
	c14 := &MockComponent{id: 14} // 14 % 8 = 6

	table.Set(c6)
	table.Set(c7)
	table.Set(c14)

	if table.Get(14) != c14 {
		t.Error("Get(14) failed (wrap around)")
	}
}

func TestComponentTable_DeleteBug_Gap(t *testing.T) {
	var table ComponentTable
	// Scenario:
	// 1: A (h=1)
	// 2: B (h=2)
	// 3: C (h=1)
	// Delete A. C should be reachable.

	c1 := &MockComponent{id: 1} // h=1
	c2 := &MockComponent{id: 2} // h=2
	c9 := &MockComponent{id: 9} // h=1

	// Add c1 (to 1)
	table.Set(c1)
	// Add c2 (to 2)
	table.Set(c2)
	// Add c9 (h=1). Slot 1 busy, Slot 2 busy. Goes to 3.
	table.Set(c9)

	// Verify setup
	if table.Get(1) != c1 { t.Fatal("Setup c1 failed") }
	if table.Get(2) != c2 { t.Fatal("Setup c2 failed") }
	if table.Get(9) != c9 { t.Fatal("Setup c9 failed") }

	// Delete 1
	table.Delete(1)

	// c2 should remain found (at 2, or moved? it shouldn't move because its h=2)
	if got := table.Get(2); got != c2 {
		t.Errorf("Get(2) = %v; want %v", got, c2)
	}

	// c9 should be found. It needs to move to 1, or just be found.
	// Since 1 is now empty (if not filled), Get(9) (h=1) will check 1. If nil, returns nil.
	if got := table.Get(9); got != c9 {
		t.Errorf("Get(9) = %v; want %v", got, c9)
	}
}
