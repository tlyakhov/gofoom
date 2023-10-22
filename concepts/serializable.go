package concepts

import (
	"reflect"
)

type ISerializable interface {
	Construct(data map[string]interface{})
	Serialize() map[string]interface{}
	SetParent(interface{})
	GetBase() *Base
	GetModel() ISerializable
}

func IndexOf(s []ISerializable, obj ISerializable) int {
	name := obj.GetBase().Name
	for i, e := range s {
		if e.GetBase().Name == name && reflect.TypeOf(obj) == reflect.TypeOf(e) {
			return i
		}
	}
	return -1
}
