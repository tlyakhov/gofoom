// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"fmt"
	"strconv"
	"strings"
)

type Entity int

func (e Entity) String() string {
	return strconv.FormatInt(int64(e), 10)
}

func ParseEntity(e string) (Entity, error) {
	v, err := strconv.ParseInt(e, 10, 32)
	return Entity(v), err
}

func (e Entity) Format(db *ECS) string {
	if e == 0 {
		return "[0] Nothing"
	}
	var sb strings.Builder

	sb.WriteString("[")
	sb.WriteString(e.String())
	sb.WriteString("] ")
	first := true
	for _, c := range db.AllComponents(e) {
		if c == nil {
			continue
		}
		if !first {
			sb.WriteString("|")
		}
		first = false
		sb.WriteString(c.String())
		/*t := reflect.TypeOf(c).Elem().String()
		split := strings.Split(t, ".")
		sb.WriteString(split[len(split)-1])*/
	}
	return sb.String()
}

func (e Entity) NameString(db *ECS) string {
	if e == 0 {
		return "0 - Nothing"
	}
	id := e.String()
	if named := GetNamed(db, e); named != nil {
		return id + " - " + named.Name
	}
	return id
}

func DeserializeEntities[T ~string | any](data []T) EntityTable {
	if data == nil {
		return nil
	}

	result := make(EntityTable, 0)
	for _, e := range data {
		switch c := any(e).(type) {
		case string:
			if entity, err := ParseEntity(c); err == nil {
				result.Set(entity)
			}
		case fmt.Stringer:
			if entity, err := ParseEntity(c.String()); err == nil {
				result.Set(entity)
			}
		}
	}
	return result
}

func ParseEntityCSV(csv string) EntityTable {
	entities := make(EntityTable, 0)
	split := strings.Split(csv, ",")
	for _, s := range split {
		if e, err := ParseEntity(strings.Trim(s, " \t\r\n")); err == nil {
			entities.Set(e)
		}
	}
	return entities
}

func ParseEntityTable(data any) EntityTable {
	var entities EntityTable
	if s, ok := data.(string); ok {
		entities = ParseEntityCSV(s)
	} else if arr, ok := data.([]string); ok {
		entities = DeserializeEntities(arr)
	} else if arr, ok := data.([]any); ok {
		entities = DeserializeEntities(arr)
	}
	return entities
}

func ParseEntitiesFromMap(data map[string]any) (EntityTable, int) {
	dataEntities := data["Entities"]
	if v, ok := data["Entity"]; ok && dataEntities == nil {
		dataEntities = v
	}
	if dataEntities == nil {
		return nil, 0
	}
	entities := ParseEntityTable(dataEntities)
	attachments := 0
	for _, e := range entities {
		if e != 0 {
			attachments++
		}
	}
	return entities, attachments
}
