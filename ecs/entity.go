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

// Entity represents an entity identifier within the ECS.
// If this is enlarged to 64 bit, then the bitmaps need to support iterating
// over larger ranges, or we need to use an entity bitmap per source
type Entity uint32

// EntitySourceID represents the identifier for the source file of an entity.
type EntitySourceID uint8

// EntityBits is the number of bits used to store the entity ID.
const EntityBits = 24

// EntitySourceIDBits is the number of bits used to store the entity source ID.
const EntitySourceIDBits = 8

// MaxEntities is the maximum number of entities that can be stored, based on
// the number of bits allocated for the entity ID.
const MaxEntities = (1 << EntityBits) - 1

/*
Entity serialization format is:
"(Delimiter)(24 bit Entity ID)[(Delimiter)(Name url encoded)(Delimiter)(8 bit
Source ID)(Delimiter)(File name url encoded)]

Why the unicode delimiter?

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
// EntityDelimiter is the delimiter used in entity serialization.
const EntityDelimiter = "∈⋮"
const entityDelimiterLength = len(EntityDelimiter)

// EntityRegexp is a regular expression used to parse entity strings.
var EntityRegexp = regexp.MustCompile(`^∈⋮(?<entity>[0-9]+)(?:∈⋮(?<name>[^∈\s]*))?(?:∈⋮(?<file_id>[0-9]+)∈⋮(?<file>[^∈\s]+))?`)

// String returns a human-readable version string representation of the entity,
// ignoring the name and original file.
func (e Entity) String() string {
	return EntityDelimiter + strconv.FormatInt(int64(e), 10)
}

// ShortString returns a concise human-readable version of the entity, including
// the source ID if it's external.
func (e Entity) ShortString() string {
	if e.IsExternal() {
		return strconv.FormatInt(int64(e&MaxEntities), 10) +
			" (" + strconv.FormatInt(int64(e>>EntityBits), 10) + ")"
	} else {
		return strconv.FormatInt(int64(e), 10)
	}
}

// SourceID returns the source ID of the entity.
func (e Entity) SourceID() EntitySourceID {
	return EntitySourceID(e >> EntityBits)
}

// IsExternal checks if the entity is external (i.e., its source ID is not 0).
func (e Entity) IsExternal() bool {
	return e.SourceID() != 0
}

// Local returns the local entity ID (excluding the source ID).
func (e Entity) Local() Entity {
	return e & MaxEntities
}

// ParseEntity parses an entity string and returns the corresponding Entity value.
func ParseEntity(e string) (Entity, error) {
	if !strings.HasPrefix(e, EntityDelimiter) {
		return 0, fmt.Errorf("Entity string `%v` should start with %v", e, EntityDelimiter)
	}
	parts := EntityRegexp.FindStringSubmatch(e)
	if parts == nil {
		return 0, errors.New("Can't parse entity " + e)
	}
	parsedEntity, err := strconv.ParseInt(parts[1], 10, EntityBits+EntitySourceIDBits)
	if len(parts) >= 4 && len(parts[3]) > 0 {
		parsedID, err := strconv.ParseInt(parts[3], 10, EntitySourceIDBits)
		if err != nil {
			return Entity(parsedEntity), err
		}
		parsedEntity |= parsedID << EntityBits
	}
	// TODO: return name
	return Entity(parsedEntity), err
}

// ParseEntityRawOrPrefixed parses an entity string that may or may not have the
// entity delimiter prefix.
func ParseEntityRawOrPrefixed(e string) (Entity, error) {
	if !strings.HasPrefix(e, EntityDelimiter) {
		v, err := strconv.ParseInt(e, 10, 64)
		return Entity(v), err
	}
	v, err := strconv.ParseInt(e[entityDelimiterLength:], 10, 64)
	return Entity(v), err
}

// Format returns a formatted string representation of the entity, including its
// name (if available) and source file (if external).
func (e Entity) Format(db *ECS) string {
	if e == 0 {
		return EntityDelimiter + "0 Nothing"
	}
	id := e.Local().String()
	if named := GetNamed(db, e); named != nil {
		return id + " " + named.Name
	}
	if e.IsExternal() {
		sourceID := e.SourceID()
		id += " (from " + db.SourceFileIDs[sourceID].Source + ")"
	}

	return id
}

// SerializeRaw serializes the entity to a string without considering the ECS
// context, allowing specifying a name and file.
func (e Entity) SerializeRaw(name string, file string) string {
	id := e.Local().String()
	if e == 0 {
		return id
	}
	if len(name) != 0 || e.IsExternal() {
		id += EntityDelimiter + url.QueryEscape(name)
	}
	if e.IsExternal() {
		id += EntityDelimiter + strconv.FormatUint(uint64(e.SourceID()), 10)
		if len(file) != 0 {
			id += EntityDelimiter + url.QueryEscape(file)
		}
	}
	return id
}

// Serialize serializes the entity to a string, including its name and source
// file information based on the ECS context.
func (e Entity) Serialize(db *ECS) string {
	id := e.Local().String()
	if e == 0 {
		return id
	}
	if named := GetNamed(db, e); named != nil {
		id += EntityDelimiter + url.QueryEscape(named.Name)
	} else if e.IsExternal() {
		id += EntityDelimiter
	}

	if e.IsExternal() {
		id += EntityDelimiter + strconv.FormatUint(uint64(e.SourceID()), 10)
		id += EntityDelimiter + url.QueryEscape(db.SourceFileIDs[e.SourceID()].Source)
	}
	return id
}

// DeserializeEntities deserializes a slice of entity strings to an EntityTable.
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

// ParseEntityCSV parses a comma-separated string of entities into an EntityTable.
func ParseEntityCSV(csv string, prefixOptional bool) EntityTable {
	entities := make(EntityTable, 0)
	split := strings.Split(csv, ",")
	fParse := ParseEntity
	if prefixOptional {
		fParse = ParseEntityRawOrPrefixed
	}
	for _, s := range split {
		trimmed := strings.Trim(s, " \t\r\n")
		if e, err := fParse(trimmed); err == nil {
			entities.Set(e)
		}
	}
	return entities
}

// ParseEntityTable parses a generic data type representing a list of entities
// into an EntityTable.
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

// ParseEntitiesFromMap extracts an EntityTable and attachment count from a map,
// handling different possible key names for the entities.
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
