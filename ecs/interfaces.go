// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import "reflect"

type Serializable interface {
	Construct(data map[string]any)
	IsSystem() bool
	Serialize() map[string]any
	// TODO: Rename to Attach
	SetECS(db *ECS)
	GetECS() *ECS
}
type Attachable interface {
	Serializable
	String() string
	IndexInECS() int
	SetIndexInECS(int)
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

// Dynamic is an interface for any value that is affected by time in the engine:
// 1. They have a lifecycle with a starting value that changes over time
// 2. They may have a "render" value interpolated between a past/future values.
type Dynamic interface {
	Serializable
	Attach(sim *Simulation)
	Detach(sim *Simulation)
	ResetToOriginal()
	RenderBlend(float64)
	NewFrame()
	GetAnimation() Animated
}
