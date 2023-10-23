package concepts

type EntityRef struct {
	components map[int]Attachable
	Entity     uint64
	DB         *EntityComponentDB
}

func (er *EntityRef) Nil() bool {
	return er.DB == nil || er.Entity == 0
}

func (er *EntityRef) Reset() {
	er.Entity = 0
	er.components = make(map[int]Attachable)
}

func (er *EntityRef) All() map[int]Attachable {
	if er.components == nil {
		c, _ := er.DB.EntityComponents.Load(er.Entity)
		er.components = c.(map[int]Attachable)
	}
	return er.components
}
func (er *EntityRef) Component(index int) Attachable {
	if index == 0 {
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
