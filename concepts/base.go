package concepts

import (
	"fmt"
	"reflect"

	"github.com/rs/xid"
	"github.com/tlyakhov/gofoom/registry"
)

type Base struct {
	ID   string   `editable:"ID"`
	Tags []string `editable:"Tags"`
}

func init() {
	registry.Instance().Register(Base{})
}

func (b *Base) GetBase() *Base {
	return b
}

func (b *Base) Initialize() {
	b.ID = xid.New().String()
}

func (b *Base) SetParent(parent interface{}) {
}

func (b *Base) Deserialize(data map[string]interface{}) {
	if b == nil {
		fmt.Printf("Error: attempting to deserialize nil *concepts.Base. Probably target type doesn't implement concepts.ISerializable.\n")
		return
	}
	b.Initialize()
	if v, ok := data["ID"]; ok {
		b.ID = v.(string)
	}
	if v, ok := data["Tags"]; ok {
		b.Tags = v.([]string)
	}
}

func MapPolyStruct(parent interface{}, data map[string]interface{}) ISerializable {
	typeMap := registry.Instance().All
	typeName := data["Type"].(string)
	//fmt.Printf("MapPolyStruct - TypeName: %v\n", typeName)
	if t, ok := typeMap[typeName]; ok {
		created := reflect.New(t).Interface()
		//fmt.Printf("MapPolyStruct - created: %v\n", reflect.ValueOf(created).Type())
		asserted := created.(ISerializable)
		//fmt.Printf("MapPolyStruct - asserted: %v\n", reflect.ValueOf(asserted).Type())
		asserted.SetParent(parent)
		asserted.Deserialize(data)
		return asserted
	}
	fmt.Printf("Warning: attempted to deserialize unknown polymorphic type: %v (onto a field of %v)\n", typeName, parent)
	return nil
}

func MapPolyArray(parent interface{}, target *[]ISerializable, data interface{}) {
	*target = make([]ISerializable, 0)
	for _, child := range data.([]map[string]interface{}) {
		item := MapPolyStruct(parent, child)
		if item == nil {
			continue
		}
		*target = append(*target, item)
	}
}

func MapArray(parent interface{}, arrayPtr interface{}, data interface{}) {
	valuePtr := reflect.ValueOf(arrayPtr)
	arrayValue := valuePtr.Elem()

	itemType := reflect.TypeOf(arrayPtr).Elem().Elem()
	arrayValue.Set(reflect.Zero(arrayValue.Type()))
	for _, child := range data.([]interface{}) {
		item := reflect.New(itemType.Elem()).Interface().(ISerializable)
		item.SetParent(parent)
		item.Deserialize(child.(map[string]interface{}))
		arrayValue.Set(reflect.Append(arrayValue, reflect.ValueOf(item)))
	}
}

func MapCollection(parent interface{}, target interface{}, data interface{}) {
	mv := reflect.ValueOf(target)
	t := mv.Type()
	mv.Elem().Set(reflect.MakeMap(t.Elem()))
	for _, child := range data.([]interface{}) {
		item := MapPolyStruct(parent, child.(map[string]interface{}))
		if item == nil {
			continue
		}
		mv.Elem().SetMapIndex(reflect.ValueOf(item.GetBase().ID), reflect.ValueOf(item))
	}
}
