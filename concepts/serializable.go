package concepts

import (
	"reflect"

	"github.com/rs/xid"
)

type ISerializable interface {
	Deserialize(data map[string]interface{})
}

type Base struct {
	ID   string   `editable:"ID" edit_type:"string"`
	Tags []string `editable:"Tags" edit_type:"tags"`
}

func (b *Base) GenerateID(target interface{}) {
	b.ID = reflect.ValueOf(target).Type().String() + "_" + xid.New().String()
}

func (b *Base) Deserialize(data map[string]interface{}) {
	b.ID = data["ID"].(string)
	b.Tags = data["Tags"].([]string)
}

func DeSeStruct(data map[string]interface{}, valid map[string]interface{}) ISerializable {
	typeName := data["Type"].(string)
	if copied, ok := valid[typeName]; ok {
		return copied.(ISerializable)
	}
	return nil
}

func DeSeArray(target *[]ISerializable, data interface{}, valid map[string]interface{}) {
	*target = make([]ISerializable, 0)
	for _, child := range data.([]map[string]interface{}) {
		*target = append(*target, DeSeStruct(child, valid))
	}
}
