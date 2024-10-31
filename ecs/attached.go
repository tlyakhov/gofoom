// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"tlyakhov/gofoom/containers"
)

// Attached represents a set of fields common to every component and implements
// the Attachable interface.
type Attached struct {
	ComponentID
	Entity        `editable:"Component" edit_type:"Component" edit_sort:"0"`
	ECS           *ECS
	System        bool // Don't serialize this entity, disallow editing
	Active        bool `editable:"Active?"`
	indexInColumn int

	// Other entities that use this component. See Linked.Entities
	linkedCopies containers.Set[Entity]
	entityStack  []Entity
}

func (a *Attached) IsActive() bool {
	return a != nil && a.Entity != 0 && a.Active
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

func (a *Attached) IsEntitySubstituted() bool {
	return len(a.entityStack) > 0
}

// Return whichever Entity this component is actually attached to (rather than
// a linked substitution)
func (a *Attached) UnlinkedEntity() Entity {
	if a.IsEntitySubstituted() {
		return a.entityStack[0]
	}
	return a.Entity
}

func (a *Attached) OnDetach() {
	if a.ECS == nil {
		return
	}
	// Remove this entity from the sources of any linked copies
	for e := range a.linkedCopies {
		if linked := GetLinked(a.ECS, e); linked != nil {
			for i, source := range linked.Sources {
				if source == e {
					linked.Sources = append(linked.Sources[:i], linked.Sources[i+1:]...)
					break
				}
			}
			linked.SourceComponents.Delete(a.ComponentID)
		}
	}
	a.linkedCopies = make(containers.Set[Entity])
	a.ECS = nil
}

func (a *Attached) IsSystem() bool {
	return a.System
}

func (a *Attached) SetColumnIndex(i int) {
	a.indexInColumn = i
}

func (a *Attached) AttachECS(db *ECS) {
	a.ECS = db
}

func (a *Attached) Construct(data map[string]any) {
	a.Active = true
	a.System = false
	a.linkedCopies = make(containers.Set[Entity])

	if data == nil {
		return
	}
	if v, ok := data["Entity"]; ok {
		a.Entity, _ = ParseEntity(v.(string))
	}
	if v, ok := data["Active"]; ok {
		a.Active = v.(bool)
	}
}

func (a *Attached) Serialize() map[string]any {
	result := map[string]any{"Entity": a.Entity.String()}
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
			result[i].AttachECS(db)
			if hook != nil {
				hook(result[i])
			}
			result[i].Construct(dataElement.(map[string]any))
		}
	} else if dataSlice, ok := data.([]map[string]any); ok {
		result = make([]PT, len(dataSlice))
		for i, dataElement := range dataSlice {
			result[i] = new(T)
			result[i].AttachECS(db)
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
