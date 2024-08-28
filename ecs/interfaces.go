// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"reflect"
	"tlyakhov/gofoom/concepts"
)

type Serializable interface {
	Construct(data map[string]any)
	IsSystem() bool
	Serialize() map[string]any
	AttachECS(db *ECS)
	GetECS() *ECS
}
type SubSerializable interface {
	Construct(ecs *ECS, data map[string]any)
	Serialize() map[string]any
}

type Attachable interface {
	Serializable
	String() string
	IndexInColumn() int
	SetColumnIndex(int)
	SetEntity(entity Entity)
	GetEntity() Entity
	OnDetach()
}

type AttachableColumn interface {
	New() Attachable
	Add(c Attachable) Attachable
	Replace(c Attachable, index int) Attachable
	Attachable(index int) Attachable
	Detach(index int)
	Type() reflect.Type
	Len() int
	String() string
}

type GenericAttachable[T any] interface {
	*T
	Attachable
}

// A DynamicType is a type constraint for anything the engine can simulate
type DynamicType interface {
	~int | ~float64 | concepts.Vector2 | concepts.Vector3 | concepts.Vector4 | concepts.Matrix2
}

// Dynamic is an interface for any value that is affected by time in the engine:
// 1. They have a lifecycle with a starting value that changes over time
// 2. They may have a "render" value interpolated between a past/future values.
type Dynamic interface {
	Serializable
	Attach(sim *Simulation)
	Detach(sim *Simulation)
	ResetToOriginal()
	Update(float64)
	Recalculate()
	NewFrame()
	GetAnimation() Animated
}
