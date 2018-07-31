package concepts

import "reflect"

type ISerializable interface {
	Initialize()
	Deserialize(data map[string]interface{})
	SetParent(interface{})
}

type Collection map[string]ISerializable

func Local(x interface{}, typeMap map[reflect.Type]reflect.Type) interface{} {
	if target, ok := typeMap[reflect.TypeOf(x)]; ok {
		return reflect.ValueOf(x).Convert(target).Interface()
	}
	return nil
}
