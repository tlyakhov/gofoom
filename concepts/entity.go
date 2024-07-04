// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"strconv"
	"strings"
)

type Entity int

func (e Entity) Format() string {
	return strconv.FormatInt(int64(e), 10)
}

func ParseEntity(e string) (Entity, error) {
	v, err := strconv.ParseInt(e, 10, 32)
	return Entity(v), err
}

func (e Entity) String(db *EntityComponentDB) string {
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

func (e Entity) NameString(db *EntityComponentDB) string {
	id := e.Format()
	if named := NamedFromDb(db, e); named != nil {
		return id + " - " + named.Name
	}
	return id
}
