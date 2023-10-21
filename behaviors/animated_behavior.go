package behaviors

import (
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/core"
)

type AnimatedBehavior struct {
	concepts.Base `editable:"^"`
	Entity        core.AbstractEntity

	Active bool `editable:"Active?"`
}

func (b *AnimatedBehavior) SetParent(parent interface{}) {
	if e, ok := parent.(core.AbstractEntity); ok {
		b.Entity = e
	} else {
		panic("Tried core.AnimatedBehavior.SetParent with a parameter that wasn't a core.AbstractEntity")
	}
}

func (b *AnimatedBehavior) Frame() {
}

func (b *AnimatedBehavior) Construct(data map[string]interface{}) {
	b.Base.Construct(data)
	b.Model = b
	b.Active = true

	if data == nil {
		return
	}

	if v, ok := data["Active"]; ok {
		b.Active = v.(bool)
	}
}

func (b *AnimatedBehavior) Serialize() map[string]interface{} {
	result := b.Base.Serialize()
	result["Active"] = b.Active
	return result
}
