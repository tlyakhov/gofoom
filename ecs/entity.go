// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"fmt"
	"strconv"
	"strings"
	"tlyakhov/gofoom/concepts"
)

type Entity int

type EntityRef[T interface {
	comparable
	Attachable
}] struct {
	Entity Entity
	Value  T
}

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

func (ref *EntityRef[T]) Refresh(db *ECS, index int) {
	var zero T
	if ref.Entity == 0 {
		ref.Value = zero
		return
	}
	if ref.Value != zero && ref.Value.GetEntity() == ref.Entity {
		return
	}
	ref.Value = db.Component(ref.Entity, index).(T)
}

func DeserializeEntities[T ~string | any](data []T) concepts.Set[Entity] {
	result := make(concepts.Set[Entity])
	if data == nil {
		return result
	}

	for _, e := range data {
		switch c := any(e).(type) {
		case string:
			if entity, err := ParseEntity(c); err == nil {
				result.Add(entity)
			}
		case fmt.Stringer:
			if entity, err := ParseEntity(c.String()); err == nil {
				result.Add(entity)
			}
		}
	}
	return result
}

func SerializeEntities(data concepts.Set[Entity]) []string {
	result := make([]string, len(data))
	i := 0
	for e := range data {
		result[i] = e.String()
		i++
	}
	return result
}
