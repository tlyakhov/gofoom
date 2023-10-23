package concepts

import (
	"github.com/rs/xid"
)

type Named struct {
	Attached
	Name string `editable:"Name"`
}

var NamedComponentIndex int

func init() {
	NamedComponentIndex = DbTypes().Register(Named{})
}

func NamedFromDb(entity *EntityRef) *Named {
	return entity.Component(NamedComponentIndex).(*Named)
}

func (n *Named) Construct(data map[string]any) {
	n.Name = xid.New().String()

	if data == nil {
		return
	}
	if v, ok := data["Name"]; ok {
		n.Name = v.(string)
	}
}

func (n *Named) Serialize() map[string]any {
	return map[string]any{"Name": n.Name}
}
