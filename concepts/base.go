package concepts

import (
	"reflect"

	"github.com/rs/xid"
)

type Base struct {
	ID   string   `editable:"ID" edit_type:"string"`
	Tags []string `editable:"Tags" edit_type:"tags"`
}

func (b *Base) Initialize() {
	b.ID = xid.New().String()
}

func (b *Base) SetParent(parent interface{}) {
}

func (b *Base) Deserialize(data map[string]interface{}) {
	b.Initialize()
	if v, ok := data["ID"]; ok {
		b.ID = v.(string)
	}
	if v, ok := data["Tags"]; ok {
		b.Tags = v.([]string)
	}
}

func (b *Base) MapPolyStruct(data map[string]interface{}, valid map[string]interface{}) ISerializable {
	typeName := data["Type"].(string)
	if copied, ok := valid[typeName]; ok {
		asserted := copied.(ISerializable)
		asserted.SetParent(b)
		asserted.Deserialize(data)
		return asserted
	}
	return nil
}

func (b *Base) MapPolyArray(target *[]ISerializable, data interface{}, valid map[string]interface{}) {
	*target = make([]ISerializable, 0)
	for _, child := range data.([]map[string]interface{}) {
		*target = append(*target, b.MapPolyStruct(child, valid))
	}
}

func (b *Base) MapArray(arrayPtr interface{}, data interface{}) {
	valuePtr := reflect.ValueOf(arrayPtr)
	arrayValue := valuePtr.Elem()

	itemType := reflect.TypeOf(arrayPtr).Elem().Elem()
	arrayValue.Set(reflect.Zero(valuePtr.Type()))
	for _, child := range data.([]map[string]interface{}) {
		item := reflect.New(itemType).Interface().(ISerializable)
		item.SetParent(b)
		item.Deserialize(child)
		arrayValue.Set(reflect.Append(arrayValue, reflect.ValueOf(item)))
	}
}

func (b *Base) MapCollection(target *Collection, data interface{}, valid map[string]interface{}) {
	*target = make(Collection, 0)
	for _, child := range data.([]interface{}) {
		item := b.MapPolyStruct(child.(map[string]interface{}), valid)
		(*target)[item.(*Base).ID] = item
	}
}
