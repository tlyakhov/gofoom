// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"reflect"
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
	SetComponentID(ComponentID)
	GetComponentID() ComponentID
	OnDetach()
}

type AttachableColumn interface {
	From(source AttachableColumn, ecs *ECS)
	New() Attachable
	Add(c Attachable) Attachable
	Replace(c Attachable, index int) Attachable
	Attachable(index int) Attachable
	Detach(index int)
	Type() reflect.Type
	Len() int
	ID() ComponentID
	String() string
}

type GenericAttachable[T any] interface {
	*T
	Attachable
}
