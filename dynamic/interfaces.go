// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0
package dynamic

import (
	"tlyakhov/gofoom/concepts"
)

// A DynamicType is a type constraint for anything the engine can simulate
type DynamicType interface {
	~int | ~float64 | concepts.Vector2 | concepts.Vector3 | concepts.Vector4 | concepts.Matrix2
}

//go:generate go run github.com/dmarkham/enumer -type=DynamicState -json
type DynamicState int

// Warning: do not change without changing Dynamic.Value to match.
const (
	Spawn DynamicState = iota
	Now
	Prev
	Render
	DynamicStates
)

// Dynamic is an interface for any value that is affected by time in the engine:
// 1. They have a lifecycle with a starting value that changes over time
// 2. They may have a "render" value interpolated between a past/future values.
type Dynamic interface {
	Spawnable
	Update(float64)
	Recalculate()
	NewFrame()
	GetAnimation() Animated
}

// Spawnable is an interface for any value that spawns at load and then
// changes over time.
type Spawnable interface {
	ResetToSpawn()
	Attach(sim *Simulation)
	Detach(sim *Simulation)
}
