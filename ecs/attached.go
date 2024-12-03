// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

// Attached represents a set of fields common to every component and implements
// the Attachable interface.
type Attached struct {
	ComponentID
	Entity
	Entities      EntityTable `editable:"Component" edit_type:"Component" edit_sort:"0"`
	Attachments   int
	ECS           *ECS
	System        bool // Don't serialize this entity, disallow editing
	Active        bool `editable:"Active?"`
	indexInColumn int
}

func (a *Attached) IsActive() bool {
	return a != nil && a.Attachments > 0 && a.Active
}

func (a *Attached) GetECS() *ECS {
	return a.ECS
}

func (a *Attached) String() string {
	return "Attached"
}

func (a *Attached) Base() *Attached {
	return a
}

func (a *Attached) MultiAttachable() bool {
	// By default, components cannot be shared
	return false
}

func (a *Attached) OnDetach(entity Entity) {
	if a.Entities.Delete(entity) {
		a.Attachments--
	}
	if a.Attachments == 1 {
		a.Entity = a.Entities.First()
	}
}

func (a *Attached) OnDelete() {
	a.ECS = nil
}

func (a *Attached) IsSystem() bool {
	return a.System
}

func (a *Attached) SetColumnIndex(i int) {
	a.indexInColumn = i
}

func (a *Attached) OnAttach(db *ECS) {
	a.ECS = db
}

func (a *Attached) Construct(data map[string]any) {
	a.Active = true
	a.System = false

	if data == nil {
		return
	}
	if v, ok := data["Active"]; ok {
		a.Active = v.(bool)
	}
	// TODO: Is this construction used anywhere? This should be happening in ECS
	//a.Entities, a.Attachments = ParseEntitiesFromMap(data)
}

func (a *Attached) Serialize() map[string]any {
	result := map[string]any{"Entities": a.Entities.Serialize()}
	if !a.Active {
		result["Active"] = a.Active
	}

	return result
}

// Confusing syntax. The constraint ensures that our underlying type has pointer
// receiver methods that implement Serializable.
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

func SerializeSlice[T Serializable](elements []T) []map[string]any {
	result := make([]map[string]any, 0, len(elements))
	for _, element := range elements {
		if element.IsSystem() {
			continue
		}
		result = append(result, element.Serialize())
	}
	return result
}
