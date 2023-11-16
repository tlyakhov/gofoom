package concepts

import (
	"reflect"
	"strconv"
	"strings"
)

type EntityRef struct {
	components []Attachable
	Entity     uint64
	DB         *EntityComponentDB
}

func (er *EntityRef) Nil() bool {
	return er == nil || er.DB == nil || er.Entity == 0
}

func (er *EntityRef) All() []Attachable {
	if er != nil && er.Entity != 0 && er.components == nil {
		er.components = er.DB.EntityComponents[er.Entity]
	}
	return er.components
}
func (er *EntityRef) Component(index int) Attachable {
	if er == nil || er.Entity == 0 || index == 0 {
		return nil
	}
	all := er.All()
	if all == nil {
		return nil
	}
	return all[index]
}

func (er *EntityRef) String() string {
	var sb strings.Builder

	sb.WriteString("[")
	sb.WriteString(strconv.FormatUint(er.Entity, 10))
	sb.WriteString("] ")
	first := true
	for _, c := range er.All() {
		if c == nil {
			continue
		}
		if !first {
			sb.WriteString("|")
		}
		first = false
		t := reflect.TypeOf(c).Elem().String()
		split := strings.Split(t, ".")
		sb.WriteString(split[len(split)-1])
	}
	return sb.String()
}

func (er *EntityRef) Serialize() string {
	return strconv.FormatUint(er.Entity, 10)
}

func SerializeEntityRefMap(data map[uint64]*EntityRef) []string {
	result := make([]string, len(data))

	i := 0
	for entity := range data {
		result[i] = strconv.FormatUint(entity, 10)
		i++
	}
	return result
}
