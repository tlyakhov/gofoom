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

// Dynamic is an interface for any value that is affected by time in the engine:
// 1. They have a lifecycle with a starting value that changes over time
// 2. They may have a "render" value interpolated between a past/future values.
type Dynamic interface {
	Attach(sim *Simulation)
	Detach(sim *Simulation)
	ResetToOriginal()
	Update(float64)
	Recalculate()
	NewFrame()
	GetAnimation() Animated
}
