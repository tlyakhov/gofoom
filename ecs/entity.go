// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// Entity represents an entity identifier within the ECS.
// If this is enlarged to 64 bit, then the bitmaps need to support iterating
// over larger ranges, or we need to use an entity bitmap per source
type Entity uint32

// EntitySourceID is the identifier for the source file of an entity.
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
"(Delimiter)(24 bit Entity ID)[(Delimiter)(Name url encoded)(Delimiter)(File hash url encoded)]

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

// EntityRegexp is a regular expression used to parse entity strings.
var EntityRegexp = regexp.MustCompile(`^∈⋮(?<entity>[0-9]+)(?:∈⋮(?<name>[^∈\s]*))?(?:∈⋮(?<file_hash>[0-9A-Fa-f]+))?`)

// EntityHumanRegexp is a regular expression used to parse entity strings
// provided by humans.
var EntityHumanRegexp = regexp.MustCompile(`\s*(?<entity>[0-9]+)\s*(?:[(](?<file_id>[0-9]+)[)])?`)

// These are indexes into regexp matches for `EntityRegexp`
const (
	EntityRegexpIdxMatch    = 0
	EntityRegexpIdxEntity   = 1
	EntityRegexpIdxName     = 2
	EntityRegexpIdxFileHash = 3
)

// These are indexes into regexp matches for `EntityHumanRegexp`
const (
	EntityHumanRegexpIdxMatch    = 0
	EntityHumanRegexpIdxEntity   = 1
	EntityHumanRegexpIdxSourceID = 2
)

// String returns a human-readable version string representation of the entity,
// ignoring the name and original file.
func (e Entity) String() string {
	if e.IsExternal() {
		return EntityDelimiter + strconv.FormatInt(int64(e.Local()), 10) +
			" (source: " + strconv.FormatInt(int64(e.SourceID()), 10) + ")"
	} else {
		return EntityDelimiter + strconv.FormatInt(int64(e), 10)
	}
}

// ShortString returns a concise human-readable version of the entity, including
// the source ID if it's external.
func (e Entity) ShortString() string {
	if e.IsExternal() {
		return strconv.FormatInt(int64(e.Local()), 10) +
			" (" + strconv.FormatInt(int64(e.SourceID()), 10) + ")"
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

func (e Entity) WithFileID(id EntitySourceID) Entity {
	return (e & MaxEntities) | Entity(id)<<EntityBits
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
	parsedEntity, err := strconv.ParseInt(parts[EntityRegexpIdxEntity], 10, EntityBits+EntitySourceIDBits)
	if len(parts) >= 4 && len(parts[EntityRegexpIdxFileHash]) > 0 {
		parsedHash, err := strconv.ParseUint(parts[EntityRegexpIdxFileHash], 16, 32)
		if err != nil {
			return Entity(parsedEntity), err
		}
		if file := SourceFileFromHash(SourceFileHash(parsedHash)); file != nil {
			parsedEntity |= int64(file.ID) << EntityBits
		} else {
			return Entity(parsedEntity), fmt.Errorf("ecs.ParseEntity: can't find ID for entity %v, hash %x", Entity(parsedEntity), parsedHash)
		}
	}
	// TODO: return name
	return Entity(parsedEntity), err
}

// ParseEntityHumanOrCanonical parses an entity string that may be in any of the
// valid forms.
func ParseEntityHumanOrCanonical(e string) (Entity, error) {
	if strings.HasPrefix(e, EntityDelimiter) {
		return ParseEntity(e)
	}
	parts := EntityHumanRegexp.FindStringSubmatch(e)
	if parts == nil {
		return 0, errors.New("Can't parse entity " + e)
	}
	parsedEntity, err := strconv.ParseInt(parts[EntityHumanRegexpIdxEntity], 10, EntityBits+EntitySourceIDBits)
	if len(parts) >= 3 && len(parts[EntityHumanRegexpIdxSourceID]) > 0 {
		parsedID, err := strconv.ParseInt(parts[EntityHumanRegexpIdxSourceID], 10, EntitySourceIDBits)
		if err != nil {
			return Entity(parsedEntity), err
		}
		parsedEntity |= parsedID << EntityBits
	}
	// TODO: return name
	return Entity(parsedEntity), err
}

// Format returns a formatted string representation of the entity, including its
// name (if available) and source file (if external).
func (e Entity) Format() string {
	if e == 0 {
		return EntityDelimiter + "0 Nothing"
	}
	id := e.Local().String()
	if named := GetNamed(e); named != nil {
		id = id + " " + named.Name
	}
	if e.IsExternal() {
		sourceID := e.SourceID()
		file := path.Base(SourceFileIDs[sourceID].Source)
		id += " (from " + file + ")"
	}

	return id
}

// SerializeRaw serializes the entity to a string with any ECS context, allowing
// specifying a name and file.
func (e Entity) SerializeRaw(name string, hash uint64) string {
	id := e.Local().String()
	if e == 0 {
		return id
	}
	if len(name) != 0 || e.IsExternal() {
		id += EntityDelimiter + url.QueryEscape(name)
	}
	if e.IsExternal() {
		id += EntityDelimiter + strconv.FormatUint(hash, 10)
	}
	return id
}

// Serialize serializes the entity to a string, including its name and source
// file information based on what was loaded.
func (e Entity) Serialize() string {
	id := EntityDelimiter + strconv.FormatUint(uint64(e.Local()), 10)
	if e == 0 {
		return id
	}
	if named := GetNamed(e); named != nil {
		id += EntityDelimiter + url.QueryEscape(named.Name)
	} else if e.IsExternal() {
		id += EntityDelimiter
	}

	if e.IsExternal() {
		hash := SourceFileIDs[e.SourceID()].Hash(true)
		id += EntityDelimiter + strconv.FormatUint(uint64(hash), 16)
	}
	return id
}

// parseEntityTableFromSlice deserializes a slice of entity strings to an EntityTable.
func parseEntityTableFromSlice[T ~string | any](data []T, humanAllowed bool) EntityTable {
	if data == nil {
		return nil
	}
	fParse := ParseEntity
	if humanAllowed {
		fParse = ParseEntityHumanOrCanonical
	}
	result := make(EntityTable, 0)
	for _, e := range data {
		switch c := any(e).(type) {
		case string:
			if entity, err := fParse(c); err == nil {
				result.Set(entity)
			} else {
				fmt.Printf("ecs.parseEntityTableFromSlice: Error %v parsing entity %v", err, e)
			}
		case fmt.Stringer:
			if entity, err := fParse(c.String()); err == nil {
				result.Set(entity)
			} else {
				fmt.Printf("ecs.parseEntityTableFromSlice: Error %v parsing entity %v", err, e)
			}
		}
	}
	return result
}

// parseEntitySliceFromSlice deserializes a slice of entity strings
func parseEntitySliceFromSlice[T ~string | any](data []T, humanAllowed bool) []Entity {
	if data == nil {
		return nil
	}
	fParse := ParseEntity
	if humanAllowed {
		fParse = ParseEntityHumanOrCanonical
	}
	result := make([]Entity, 0)
	for _, e := range data {
		switch c := any(e).(type) {
		case string:
			if entity, err := fParse(c); err == nil {
				result = append(result, entity)
			} else {
				fmt.Printf("ecs.parseEntitySliceFromSlice: Error %v parsing entity %v", err, e)
			}
		case fmt.Stringer:
			if entity, err := fParse(c.String()); err == nil {
				result = append(result, entity)
			} else {
				fmt.Printf("ecs.parseEntitySliceFromSlice: Error %v parsing entity %v", err, e)
			}
		}
	}
	return result
}

// parseEntityTableCSV parses a comma-separated string of entities into an EntityTable.
func parseEntityTableCSV(csv string, humanAllowed bool) EntityTable {
	entities := make(EntityTable, 0)
	fParse := ParseEntity
	if humanAllowed {
		fParse = ParseEntityHumanOrCanonical
	}
	for s := range strings.SplitSeq(csv, ",") {
		trimmed := strings.TrimSpace(s)
		if e, err := fParse(trimmed); err == nil {
			entities.Set(e)
		}
	}
	return entities
}

// parseEntitySliceCSV parses a comma-separated string of entities
func parseEntitySliceCSV(csv string, humanAllowed bool) []Entity {
	entities := make([]Entity, 0)
	fParse := ParseEntity
	if humanAllowed {
		fParse = ParseEntityHumanOrCanonical
	}
	for s := range strings.SplitSeq(csv, ",") {
		trimmed := strings.TrimSpace(s)
		if e, err := fParse(trimmed); err == nil {
			entities = append(entities, e)
		}
	}
	return entities
}

// ParseEntityTable parses a generic data type representing a list of entities
// into an EntityTable.
func ParseEntityTable(data any, humanAllowed bool) EntityTable {
	var entities EntityTable
	if s, ok := data.(string); ok {
		entities = parseEntityTableCSV(s, humanAllowed)
	} else if arr, ok := data.([]string); ok {
		entities = parseEntityTableFromSlice(arr, humanAllowed)
	} else if arr, ok := data.([]any); ok {
		entities = parseEntityTableFromSlice(arr, humanAllowed)
	}
	return entities
}

// ParseEntitySlice parses a generic data type representing a list of entities
func ParseEntitySlice(data any, humanAllowed bool) []Entity {
	var entities []Entity
	if s, ok := data.(string); ok {
		entities = parseEntitySliceCSV(s, humanAllowed)
	} else if arr, ok := data.([]string); ok {
		entities = parseEntitySliceFromSlice(arr, humanAllowed)
	} else if arr, ok := data.([]any); ok {
		entities = parseEntitySliceFromSlice(arr, humanAllowed)
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
	entities := ParseEntityTable(dataEntities, false)
	attachments := 0
	for _, e := range entities {
		if e != 0 {
			attachments++
		}
	}
	return entities, attachments
}
