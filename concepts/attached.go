// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"strings"
)

type Attached struct {
	Entity                  `editable:"Component" edit_type:"Component" edit_sort:"0"`
	DB                      *EntityComponentDB
	Active                  bool `editable:"Active?"`
	ActiveWhileEditorPaused bool `editable:"Active when editor paused?"`
	indexInDB               int
}

type Attachable interface {
	Serializable
	String() string
	IndexInDB() int
	SetIndexInDB(int)
	SetEntity(entity Entity)
	GetEntity() Entity
}

type Serializable interface {
	Construct(data map[string]any)
	Serialize() map[string]any
	// TODO: Rename to Attach
	SetDB(db *EntityComponentDB)
	GetDB() *EntityComponentDB
}

var AttachedComponentIndex int

func init() {
	AttachedComponentIndex = DbTypes().Register(Attached{}, nil)
}

func (a *Attached) IsActive() bool {
	return a != nil && a.Entity != 0 && a.Active && (a.ActiveWhileEditorPaused || !a.DB.Simulation.EditorPaused)
}

func (a *Attached) String() string {
	return "Attached"
}

func (a *Attached) IndexInDB() int {
	return a.indexInDB
}

func (a *Attached) SetIndexInDB(i int) {
	a.indexInDB = i
}

func (a *Attached) SetDB(db *EntityComponentDB) {
	a.DB = db
}

func (a *Attached) GetDB() *EntityComponentDB {
	return a.DB
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
	result := map[string]any{"Entity": a.Entity.Format()}
	if !a.Active {
		result["Active"] = a.Active
	}
	if !a.ActiveWhileEditorPaused {
		result["ActiveWhileEditorPaused"] = a.ActiveWhileEditorPaused
	}

	return result
}

func (a *Attached) DeserializeComponentList(list *map[int]bool, name string, data map[string]any) {
	v, ok := data[name]
	if !ok {
		return
	}
	listString, ok := v.(string)
	if !ok {
		return
	}
	split := strings.Split(listString, ",")
	*list = make(map[int]bool)
	for _, typeName := range split {
		componentIndex := DbTypes().Indexes[typeName]
		if componentIndex != 0 {
			(*list)[componentIndex] = true
		}
	}
}

func (a *Attached) SerializeComponentList(list map[int]bool, name string, result map[string]any) {
	if len(list) == 0 {
		return
	}

	types := make([]string, 0)
	for index := range list {
		types = append(types, DbTypes().Types[index].String())
	}
	result[name] = strings.Join(types, ",")
}

// Confusing syntax. The constraint ensures that our underlying type has pointer
// receiver methods that implement Serializable.
func ConstructSlice[PT interface {
	*T
	Serializable
}, T any](db *EntityComponentDB, data any) []PT {
	var result []PT

	if dataSlice, ok := data.([]any); ok {
		result = make([]PT, len(dataSlice))
		for i, dataElement := range dataSlice {
			result[i] = new(T)
			result[i].SetDB(db)
			result[i].Construct(dataElement.(map[string]any))
		}
	}
	return result
}

func SerializeSlice[T Serializable](elements []T) []map[string]any {
	result := make([]map[string]any, len(elements))
	for i, element := range elements {
		result[i] = element.Serialize()
	}
	return result
}
