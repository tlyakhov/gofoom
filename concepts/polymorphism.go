package concepts

import "reflect"

func MapArray(parent any, arrayPtr any, data any) {
	valuePtr := reflect.ValueOf(arrayPtr)
	arrayValue := valuePtr.Elem()

	itemType := reflect.TypeOf(arrayPtr).Elem().Elem()
	arrayValue.Set(reflect.Zero(arrayValue.Type()))
	for _, child := range data.([]any) {
		item := reflect.New(itemType.Elem()).Interface().(Attachable)
		//item.SetParent(parent)
		item.Construct(child.(map[string]any))
		arrayValue.Set(reflect.Append(arrayValue, reflect.ValueOf(item)))
	}
}

/*func MapPolyStruct(parent any, data map[string]any) IComponent {
	typeMap := concepts.DB().AllTypes
	typeName := data["Type"].(string)
	//fmt.Printf("MapPolyStruct - TypeName: %v\n", typeName)
	if t, ok := typeMap[typeName]; ok {
		created := reflect.New(t).Interface()
		//fmt.Printf("MapPolyStruct - created: %v\n", reflect.ValueOf(created).Type())
		asserted := created.(IComponent)
		//fmt.Printf("MapPolyStruct - asserted: %v\n", reflect.ValueOf(asserted).Type())
		asserted.SetParent(parent)
		asserted.Construct(data)
		return asserted
	}
	fmt.Printf("Warning: attempted to deserialize unknown polymorphic type: %v (onto a field of %v)\n", typeName, parent)
	return nil
}

func MapPolyArray(parent any, target *[]IComponent, data any) {
	*target = make([]IComponent, 0)
	for _, child := range data.([]map[string]any) {
		item := MapPolyStruct(parent, child)
		if item == nil {
			continue
		}
		*target = append(*target, item)
	}
}

func MapCollection(parent any, target any, data any) {
	mv := reflect.ValueOf(target)
	t := mv.Type()
	mv.Elem().Set(reflect.MakeMap(t.Elem()))
	for _, child := range data.([]any) {
		item := MapPolyStruct(parent, child.(map[string]any))
		if item == nil {
			continue
		}
		mv.Elem().SetMapIndex(reflect.ValueOf(item.GetEntity().Name), reflect.ValueOf(item))
	}
}
*/
