// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

// ComponentFlags represents flags that can be associated with a component.
//go:generate go run github.com/dmarkham/enumer -type=ComponentFlags -json
type ComponentFlags int

const (
	// ComponentNoSave indicates that the component should not be saved to disk.
	ComponentNoSave ComponentFlags = 1 << iota
	// ComponentHideInEditor indicates that the component should be hidden in the editor.
	ComponentHideInEditor
	// ComponentLockedInEditor indicates that the component should be locked in the editor, preventing modifications.
	ComponentLockedInEditor
)

// ComponentInternal is a combination of flags indicating that the component is internal and should not be saved or modified by the user.
const ComponentInternal = ComponentNoSave | ComponentHideInEditor | ComponentLockedInEditor

// Attached represents a set of fields common to every component and implements
// the Attachable interface. It provides basic functionality for managing components within the ECS.
type Attached struct {
	// ComponentID is the unique identifier for the component type.
	ComponentID
	// Entity is the ID of the primary entity to which this component is attached.
	Entity
	// Entities is a table of entities to which this component is attached. This is used for components that can be attached to multiple entities.
	Entities EntityTable `editable:"Component" edit_type:"Component" edit_sort:"0"`
	// Active indicates whether the component is active. Inactive components are not processed by controllers.
	Active bool `editable:"Active?"`
	// Attachments is a reference counter tracking the number of entities this component is attached to.
	Attachments int // Reference counter
	// ECS is a pointer to the ECS instance that manages this component.
	ECS *ECS
	// Flags are bit flags that control the behavior of the component, such as whether it is saved or visible in the editor.
	Flags ComponentFlags
	// indexInColumn is the index of this component within its column in the ECS.
	indexInColumn int
}

// IsActive checks if the component is active, attached to any entities, and not nil.
func (a *Attached) IsActive() bool {
	return a != nil && a.Attachments > 0 && a.Active
}

// GetECS returns the ECS instance associated with this component.
func (a *Attached) GetECS() *ECS {
	return a.ECS
}

// String returns a string representation of the Attached struct.
func (a *Attached) String() string {
	return "Attached"
}

// Base returns a pointer to the base Attached struct.
func (a *Attached) Base() *Attached {
	return a
}

// MultiAttachable returns whether this component type can be attached to multiple entities. By default, components cannot be shared.
func (a *Attached) MultiAttachable() bool {
	// By default, components cannot be shared
	return false
}

// OnDetach is called when the component is detached from an entity. It updates the attachments counter and the primary entity.
func (a *Attached) OnDetach(entity Entity) {
	if a.Entities.Delete(entity) {
		a.Attachments--
	}
	a.Entity = a.Entities.First()
}

// OnDelete is called when the component is deleted from the ECS. It clears the ECS pointer.
func (a *Attached) OnDelete() {
	a.ECS = nil
}

// SetColumnIndex sets the index of this component within its column.
func (a *Attached) SetColumnIndex(i int) {
	a.indexInColumn = i
}

// OnAttach is called when the component is attached to an ECS. It sets the ECS pointer.
func (a *Attached) OnAttach(db *ECS) {
	a.ECS = db
}

// Construct initializes the component with data from a map. It sets the active flag to true and applies any provided data.
func (a *Attached) Construct(data map[string]any) {
	a.Active = true
	a.Flags = 0

	if data == nil {
		return
	}
	if v, ok := data["Active"]; ok {
		a.Active = v.(bool)
	}
	// TODO: Is this construction used anywhere? This should be happening in ECS
	//a.Entities, a.Attachments = ParseEntitiesFromMap(data)
}

// Serialize returns a map representing the component's data for serialization. It includes the entities and the active flag if it's false.
func (a *Attached) Serialize() map[string]any {
	result := map[string]any{"Entities": a.Entities.Serialize(a.ECS)}
	if !a.Active {
		result["Active"] = a.Active
	}

	return result
}

// ConstructSlice constructs a slice of components from a slice of data maps.
// The type parameter PT must be a pointer to a type T that implements the Serializable interface.
// An optional hook function can be provided to perform additional initialization on each component.
func ConstructSlice[PT interface {
	*T
	Serializable
}, T any](db *ECS, data any, hook func(item PT)) []PT {
	var result []PT

	if dataSlice, ok := data.([]any); ok {
		result = make([]PT, len(dataSlice))
		for i, dataElement := range dataSlice {
			result[i] = new(T)
			result[i].OnAttach(db)
			if hook != nil {
				hook(result[i])
			}
			result[i].Construct(dataElement.(map[string]any))
		}
	} else if dataSlice, ok := data.([]map[string]any); ok {
		result = make([]PT, len(dataSlice))
		for i, dataElement := range dataSlice {
			result[i] = new(T)
			result[i].OnAttach(db)
			if hook != nil {
				hook(result[i])
			}
			result[i].Construct(dataElement)
		}
	}
	return result
}

// SerializeSlice serializes a slice of Serializable components into a slice of maps.
// It skips components that have the ComponentNoSave flag set.
func SerializeSlice[T Serializable](elements []T) []map[string]any {
	result := make([]map[string]any, 0, len(elements))
	for _, element := range elements {
		// Attachables can have a flag to not serialize them
		if a, ok := any(element).(Attachable); ok {
			if a.Base().Flags&ComponentNoSave != 0 {
				continue
			}
		}
		result = append(result, element.Serialize())
	}
	return result
}
