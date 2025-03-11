// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// If this is enlarged to 64 bit, then the bitmaps need to support iterating
// over larger ranges, or we need to use an entity bitmap per source
type Entity uint32
type EntitySourceID uint8

const EntityBits = 24
const EntitySourceIDBits = 8
const MaxEntities = (1 << EntityBits) - 1

/*
Entity serialization format is:
"(Prefix)(16 bit Source ID + 48 bit Entity ID)[(Prefix)Name (url encoded)

Why the unicode prefix?

	0x2208 ELEMENT OF
	0x22EE VERTICAL ELLIPSIS

Because it gives us a few benefits:
 1. This sequence is unlikely to be used in some other context. It's a
    human- and machine-readable way to identify entities when
    serializing/deserializing.
 2. We can do complex transformations on serialized data, even when the
    file format is untyped (e.g. YAML/JSON). For example, we can create
    "common" data files, like prefabs, and #include them in other files,
    and have the ECS intelligently map the entity IDs across file boundaries.
 3. It forces target systems to be UTF-8 compliant - serialization will be
    entirely broken otherwise.
*/
const EntityPrefix = "∈⋮"
const entityPrefixLength = len(EntityPrefix)

var EntityRegexp = regexp.MustCompile(`^∈⋮([0-9]+)(?:∈⋮([^∈ \r\n\t]+)(?:∈⋮([^∈ \r\n\t]+))?)?`)

// Ignores name/original file
func (e Entity) String() string {
	return EntityPrefix + strconv.FormatInt(int64(e), 10)
}

func (e Entity) SourceID() EntitySourceID {
	return EntitySourceID(e >> EntityBits)
}

func (e Entity) ExcludeSourceID() Entity {
	return e & MaxEntities
}

func (e Entity) IsExternal() bool {
	return e.SourceID() != 0
}

func (e Entity) Local() Entity {
	return e & MaxEntities
}

func ParseEntity(e string) (Entity, error) {
	if !strings.HasPrefix(e, EntityPrefix) {
		return 0, fmt.Errorf("Entity string `%v` should start with %v", e, EntityPrefix)
	}
	parts := EntityRegexp.FindStringSubmatch(e)
	if parts == nil {
		return 0, errors.New("Can't parse entity " + e)
	}
	v, err := strconv.ParseInt(parts[1], 10, 64)
	// TODO: return name
	return Entity(v), err
}

func ParseEntityPrefixOptional(e string) (Entity, error) {
	if !strings.HasPrefix(e, EntityPrefix) {
		v, err := strconv.ParseInt(e, 10, 64)
		return Entity(v), err
	}
	v, err := strconv.ParseInt(e[entityPrefixLength:], 10, 64)
	return Entity(v), err
}

func (e Entity) Format(db *ECS) string {
	if e == 0 {
		return EntityPrefix + "0 Nothing"
	}
	id := e.String()
	if named := GetNamed(db, e); named != nil {
		return id + " " + named.Name
	}

	return id
}

func (e Entity) SerializeRaw(name string, file string) string {
	id := e.String()
	if e == 0 {
		return id
	}
	if len(name) != 0 {
		id += EntityPrefix + url.QueryEscape(name)
	}
	if len(file) != 0 {
		id += EntityPrefix + url.QueryEscape(file)
	}
	return id
}

func (e Entity) Serialize(db *ECS) string {
	id := e.String()
	if e == 0 {
		return id
	}
	if named := GetNamed(db, e); named != nil {
		id += EntityPrefix + url.QueryEscape(named.Name)
	}
	if e.IsExternal() {
		id += EntityPrefix + url.QueryEscape(db.SourceFileIDs[e.SourceID()].Source)
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
			} else {
				fmt.Printf("ecs.DeserializeEntities: Error %v parsing entity %v", err, e)
			}
		case fmt.Stringer:
			if entity, err := ParseEntity(c.String()); err == nil {
				result.Set(entity)
			} else {
				fmt.Printf("ecs.DeserializeEntities: Error %v parsing entity %v", err, e)
			}
		}
	}
	return result
}

func ParseEntityCSV(csv string, prefixOptional bool) EntityTable {
	entities := make(EntityTable, 0)
	split := strings.Split(csv, ",")
	fParse := ParseEntity
	if prefixOptional {
		fParse = ParseEntityPrefixOptional
	}
	for _, s := range split {
		trimmed := strings.Trim(s, " \t\r\n")
		if e, err := fParse(trimmed); err == nil {
			entities.Set(e)
		}
	}
	return entities
}

func ParseEntityTable(data any) EntityTable {
	var entities EntityTable
	if s, ok := data.(string); ok {
		entities = ParseEntityCSV(s, false)
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
