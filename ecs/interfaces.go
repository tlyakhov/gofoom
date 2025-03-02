// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"reflect"
)

type Serializable interface {
	Construct(data map[string]any)
	Serialize() map[string]any
	OnAttach(db *ECS)
	GetECS() *ECS
}
type SubSerializable interface {
	Construct(ecs *ECS, data map[string]any)
	Serialize() map[string]any
}

type Attachable interface {
	Serializable
	String() string
	OnDetach(Entity)
	OnDelete()
	IsActive() bool
	MultiAttachable() bool
	Base() *Attached
}

type AttachableColumn interface {
	From(source AttachableColumn, ecs *ECS)
	New() Attachable
	Add(c *Attachable)
	Replace(c *Attachable, index int)
	Attachable(index int) Attachable
	Detach(index int)
	Type() reflect.Type
	Len() int
	Cap() int
	ID() ComponentID
	String() string
}

type GenericAttachable[T any] interface {
	*T
	Attachable
}

type ControllerMethod uint32

const (
	ControllerAlways ControllerMethod = 1 << iota
	ControllerRecalculate
)

type Controller interface {
	ComponentID() ComponentID
	Methods() ControllerMethod
	EditorPausedMethods() ControllerMethod
	// Return false if controller shouldn't run for this entity
	Target(Attachable, Entity) bool
	Always()
	Recalculate()
}
