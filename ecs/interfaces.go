// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"reflect"
)

// Serializable is an interface for components that can be serialized and deserialized.
type Serializable interface {
	// Construct initializes the component from a map of data.
	Construct(data map[string]any)
	// Serialize returns a map representing the component's data for serialization.
	Serialize() map[string]any
	// OnAttach is called when the component is attached to an ECS instance.
	OnAttach(db *ECS)
	// GetECS returns the ECS instance associated with the component.
	GetECS() *ECS
}

// SubSerializable is an interface for sub-components that can be serialized and deserialized.
type SubSerializable interface {
	// Construct initializes the sub-component from a map of data and an ECS instance.
	Construct(ecs *ECS, data map[string]any)
	// Serialize returns a map representing the sub-component's data for serialization.
	Serialize() map[string]any
}

// Attachable is an interface for components that can be attached to entities in the ECS.
type Attachable interface {
	Serializable
	// String returns a string representation of the component.
	String() string
	// OnDetach is called when the component is detached from an entity.
	OnDetach(Entity)
	// OnDelete is called when the component is deleted from the ECS.
	OnDelete()
	// IsActive checks if the component is active.
	IsActive() bool
	// MultiAttachable returns whether this component type can be attached to multiple entities.
	MultiAttachable() bool
	// Base returns a pointer to the base Attached struct.
	Base() *Attached
}

// AttachableColumn is an interface for managing a column of attachable components of a specific type.
type AttachableColumn interface {
	// From initializes a column from another column of the same type.
	From(source AttachableColumn, ecs *ECS)
	// New creates a new Attachable component of the type stored in this column.
	New() Attachable
	// Add adds a component to the column.
	Add(c *Attachable)
	// Replace replaces the component at the given index with the provided component.
	Replace(c *Attachable, index int)
	// Attachable retrieves the component at the given index as an Attachable interface.
	Attachable(index int) Attachable
	// Detach removes the component at the given index from the column.
	Detach(index int)
	// Type returns the reflect.Type of the component data stored in this column.
	Type() reflect.Type
	// Len returns the number of components currently stored in this column.
	Len() int
	// Cap returns the total capacity of this column.
	Cap() int
	// ID returns the component ID associated with this column.
	ID() ComponentID
	// String returns a string representation of the component type stored in this column.
	String() string
}

// GenericAttachable is a generic interface constraint for types that can be attached as components.
type GenericAttachable[T any] interface {
	*T
	Attachable
}

// ControllerMethod represents a bitmask of methods that a controller can implement.
type ControllerMethod uint32

const (
	// ControllerAlways indicates that the controller's Always method should be called every tick.
	ControllerAlways ControllerMethod = 1 << iota
	// ControllerRecalculate indicates that the controller's Recalculate method should be called when a component is attached or detached.
	ControllerRecalculate
)

// Controller is an interface for defining controllers that act on components within the ECS.
type Controller interface {
	// ComponentID returns the ID of the component type that this controller operates on.
	ComponentID() ComponentID
	// Methods returns a bitmask of the controller methods that this controller implements.
	Methods() ControllerMethod
	// EditorPausedMethods returns a bitmask of the controller methods that this controller implements when the editor is paused.
	EditorPausedMethods() ControllerMethod
	// Target determines whether the controller should act on a specific entity and component.
	// Return false if controller shouldn't run for this entity
	Target(Attachable, Entity) bool
	// Always is called every tick for entities that match the controller's component ID and target criteria.
	Always()
	// Recalculate is called when a component is attached or detached, or when a linked component changes.
	Recalculate()
}
