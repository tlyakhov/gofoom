package concepts

type ISerializable interface {
	Initialize()
	Deserialize(data map[string]interface{})
	SetParent(interface{})
	GetBase() *Base
}
