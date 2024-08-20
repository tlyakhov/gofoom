// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

type Attached struct {
	Entity                  `editable:"Component" edit_type:"Component" edit_sort:"0"`
	ECS                     *ECS
	System                  bool // Don't serialize this entity, disallow editing
	Active                  bool `editable:"Active?"`
	ActiveWhileEditorPaused bool `editable:"Active when editor paused?"`
	indexInECS              int
}

var AttachedComponentIndex int

func init() {
	AttachedComponentIndex = Types().Register(Attached{}, nil)
}

func (a *Attached) IsActive() bool {
	return a != nil && a.Entity != 0 && a.Active && (a.ActiveWhileEditorPaused || !a.ECS.Simulation.EditorPaused)
}

func (a *Attached) String() string {
	return "Attached"
}

func (a *Attached) IndexInECS() int {
	return a.indexInECS
}

func (a *Attached) OnDetach() {
}

func (a *Attached) IsSystem() bool {
	return a.System
}

func (a *Attached) SetIndexInECS(i int) {
	a.indexInECS = i
}

func (a *Attached) SetECS(db *ECS) {
	a.ECS = db
}

func (a *Attached) GetECS() *ECS {
	return a.ECS
}

func (a *Attached) SetEntity(entity Entity) {
	a.Entity = entity
}

func (a *Attached) GetEntity() Entity {
	return a.Entity
}

func (a *Attached) Construct(data map[string]any) {
	a.Active = true
	a.ActiveWhileEditorPaused = true

	if data == nil {
		return
	}
	if v, ok := data["Entity"]; ok {
		a.Entity, _ = ParseEntity(v.(string))
	}
	if v, ok := data["Active"]; ok {
		a.Active = v.(bool)
	}
	if v, ok := data["ActiveWhileEditorPaused"]; ok {
		a.ActiveWhileEditorPaused = v.(bool)
	}
}

func (a *Attached) Serialize() map[string]any {
	result := map[string]any{"Entity": a.Entity.String()}
	if !a.Active {
		result["Active"] = a.Active
	}
	if !a.ActiveWhileEditorPaused {
		result["ActiveWhileEditorPaused"] = a.ActiveWhileEditorPaused
	}

	return result
}

// Confusing syntax. The constraint ensures that our underlying type has pointer
// receiver methods that implement Serializable.
func ConstructSlice[PT interface {
	*T
	Serializable
}, T any](db *ECS, data any) []PT {
	var result []PT

	if dataSlice, ok := data.([]any); ok {
		result = make([]PT, len(dataSlice))
		for i, dataElement := range dataSlice {
			result[i] = new(T)
			result[i].SetECS(db)
			result[i].Construct(dataElement.(map[string]any))
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
