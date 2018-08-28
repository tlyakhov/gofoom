package concepts

import (
	"reflect"
)

type ISerializable interface {
	Initialize()
	Deserialize(data map[string]interface{})
	Serialize() map[string]interface{}
	SetParent(interface{})
	GetBase() *Base
}

func IndexOf(s []ISerializable, obj ISerializable) int {
	id := obj.GetBase().ID
	for i, e := range s {
		if e.GetBase().ID == id && reflect.TypeOf(obj) == reflect.TypeOf(e) {
			return i
		}
	}
	return -1
}
