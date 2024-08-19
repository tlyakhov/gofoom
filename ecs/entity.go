// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"strconv"
	"strings"
)

type Entity int

type EntityRef[T interface {
	comparable
	Attachable
}] struct {
	Entity Entity
	Value  T
}

func (e Entity) Format() string {
	return strconv.FormatInt(int64(e), 10)
}

func ParseEntity(e string) (Entity, error) {
	v, err := strconv.ParseInt(e, 10, 32)
	return Entity(v), err
}

func (e Entity) String(db *ECS) string {
	if e == 0 {
		return "[0] Nothing"
	}
	var sb strings.Builder

	sb.WriteString("[")
	sb.WriteString(e.Format())
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
	id := e.Format()
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
