package concepts

type EntityRef struct {
	components []Attachable
	Entity     uint64
	DB         *EntityComponentDB
}

func (er *EntityRef) Nil() bool {
	return er == nil || er.DB == nil || er.Entity == 0
}

func (er *EntityRef) Reset() {
	if er == nil {
		return
	}
	er.Entity = 0
	er.components = nil
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
	return er.All()[index]
}

func DeserializeEntityRefs(data []any) map[uint64]EntityRef {
	result := make(map[uint64]EntityRef)

	for _, v := range data {
		entity := v.(uint64)
		result[entity] = EntityRef{Entity: entity}
	}
	return result
}

func SerializeEntityRefMap(data map[uint64]EntityRef) []uint64 {
	result := make([]uint64, len(data))

	i := 0
	for entity := range data {
		result[i] = entity
		i++
	}
	return result
}
