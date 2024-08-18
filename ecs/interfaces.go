// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

type Attachable interface {
	Serializable
	String() string
	IndexInDB() int
	SetIndexInDB(int)
	SetEntity(entity Entity)
	GetEntity() Entity
	OnDetach()
}

type Serializable interface {
	Construct(data map[string]any)
	IsSystem() bool
	Serialize() map[string]any
	// TODO: Rename to Attach
	SetDB(db *ECS)
	GetDB() *ECS
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
