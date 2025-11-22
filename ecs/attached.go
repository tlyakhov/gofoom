// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"tlyakhov/gofoom/concepts"

	"github.com/spf13/cast"
)

// There are architectural tradeoffs here. The whole point of the ECS is to have
// all the data for a given component be next to each other in memory and enable
// efficient access by controllers. However, with the `Attached` mixin, we lose
// that by having to include all these extra fields. On top of that, accessing
// those fields via interface is wasteful and makes it much harder to optimize.
// That said, the fields here are 8+2+2+4+8=24 bytes, so maybe it's not too
// bad.
// Is the right approach to store this extra data separately? maybe

// Attached has a set of fields common to every component and implements
// the Component interface. It is required for every component in the ECS.
type Attached struct {
	// Entity is the ID of the primary entity to which this component is attached.
	Entity
	// Attachments is a reference counter tracking the number of entities this
	// component is attached to.
	Attachments uint16
	// Flags are bit flags that control the behavior of the component, such as
	// whether it is saved or visible in the editor.
	Flags ComponentFlags `editable:"Flags" edit_type:"Flags"`
	// indexInArena is the index of this component within its arena in the ECS.
	indexInArena int
	// Entities is a table of entities to which this component is attached. This
	// is used for components that can be attached to multiple entities.
	Entities EntityTable `editable:"Component" edit_type:"Component" edit_sort:"0"`
}

// IsActive checks if the component is active, attached to any entities, and not nil.
func (a *Attached) IsActive() bool {
	return a.IsAttached() && (a.Flags&ComponentActive != 0)
}

// IsAttached checks if the component is active, attached to any entities, and not nil.
func (a *Attached) IsAttached() bool {
	return a != nil && a.Attachments > 0
}

// IsExternal checks if the component is sourced from another file.
func (a *Attached) IsExternal() bool {
	if a.Entity.IsExternal() {
		return true
	}

	for _, e := range a.Entities {
		if e != 0 && e.IsExternal() {
			return true
		}
	}
	return false
}

// ExternalEntities returns an array of external file entities
func (a *Attached) ExternalEntities() []Entity {
	if a.Entity.IsExternal() {
		return []Entity{a.Entity}
	}

	result := []Entity{}
	for _, e := range a.Entities {
		if e != 0 && e.IsExternal() {
			result = append(result, e)
		}
	}
	return result
}

// String returns a string representation - helpful in the editor or debugging
func (a *Attached) String() string {
	return "Attached"
}

// Base returns a pointer to the base Attached struct. Used all over the place
// to access fields using the Attachable interface.
func (a *Attached) Base() *Attached {
	return a
}

// MultiAttachable returns whether this component type can be attached to
// multiple entities. By default, components cannot be shared.
func (a *Attached) MultiAttachable() bool {
	// By default, components cannot be shared
	return false
}

// OnDetach is called when the component is detached from an entity. It updates
// the attachments counter and the primary entity.
func (a *Attached) OnDetach(entity Entity) {
	if a.Entities.Delete(entity) {
		a.Attachments--
	}
	a.Entity = a.Entities.First()
}

// OnDelete is called when the component is deleted from the arena.
func (a *Attached) OnDelete() {
	a.Attachments = 0
}

// SetArenaIndex sets the index of this component within its arena.
func (a *Attached) SetArenaIndex(i int) {
	a.indexInArena = i
}

// OnAttach is called when the component is attached.
func (a *Attached) OnAttach() {
}

// Construct initializes the component with data from a map. It sets the active
// flag to true and applies any provided data.
func (a *Attached) Construct(data map[string]any) {
	a.Flags = ComponentActive

	if data == nil {
		return
	}
	if v, ok := data["_Flags"]; ok {
		a.Flags = concepts.ParseFlags(cast.ToString(v), ComponentFlagsString)
	}
	// TODO: Is this construction used anywhere? This should be happening outside
	//a.Entities, a.Attachments = ParseEntitiesFromMap(data)
}

// Serialize returns a map representing the component's data for serialization.
// It includes the entities and the active flag if it's false.
func (a *Attached) Serialize() map[string]any {
	result := map[string]any{"Entities": a.Entities.Serialize()}
	if a.Flags != ComponentActive {
		result["_Flags"] = concepts.SerializeFlags(a.Flags, ComponentFlagsValues())
	}

	return result
}

// ConstructSlice constructs a slice of components from a slice of data maps.
// The type parameter PT must be a pointer to a type T that implements the Serializable interface.
// An optional hook function can be provided to perform additional initialization on each component.
func ConstructSlice[PT interface {
	*T
	Serializable
}, T any](data any, hook func(item PT)) []PT {
	var result []PT

	if dataSlice, ok := data.([]any); ok {
		result = make([]PT, len(dataSlice))
		for i, dataElement := range dataSlice {
			result[i] = new(T)
			if u, ok := any(result[i]).(Attachable); ok {
				u.OnAttach()
			}
			if hook != nil {
				hook(result[i])
			}
			result[i].Construct(dataElement.(map[string]any))
		}
	} else if dataSlice, ok := data.([]map[string]any); ok {
		result = make([]PT, len(dataSlice))
		for i, dataElement := range dataSlice {
			result[i] = new(T)
			if u, ok := any(result[i]).(Attachable); ok {
				u.OnAttach()
			}
			if hook != nil {
				hook(result[i])
			}
			result[i].Construct(dataElement)
		}
	}
	return result
}

// SerializeSlice turns a slice of Serializable components into a slice of maps
// suitable for yaml/json. It skips components that have the ComponentNoSave flag set.
func SerializeSlice[T Serializable](elements []T) []map[string]any {
	result := make([]map[string]any, 0, len(elements))
	for _, element := range elements {
		// Attachables can have a flag to not serialize them
		if a, ok := any(element).(Component); ok {
			if a.Base().Flags&ComponentNoSave != 0 {
				continue
			}
		}
		result = append(result, element.Serialize())
	}
	return result
}
