package core

import (
	"github.com/tlyakhov/gofoom/concepts"
)

type AnimatedBehavior struct {
	concepts.Base `editable:"^"`
	Entity        AbstractEntity

	Active bool `editable:"Active?"`
}

func (b *AnimatedBehavior) Initialize() {
	b.Base.Initialize()
	b.Active = true
}

func (b *AnimatedBehavior) Animated() *AnimatedBehavior {
	return b
}

func (b *AnimatedBehavior) SetParent(parent interface{}) {
	if e, ok := parent.(AbstractEntity); ok {
		b.Entity = e
	} else {
		panic("Tried core.AnimatedBehavior.SetParent with a parameter that wasn't a core.AbstractEntity")
	}
}

func (b *AnimatedBehavior) Frame(lastFrameTime float64) {
}

func (b *AnimatedBehavior) Deserialize(data map[string]interface{}) {
	b.Initialize()
	b.Base.Deserialize(data)

	if v, ok := data["Active"]; ok {
		b.Active = v.(bool)
	}
}

func (b *AnimatedBehavior) Serialize() map[string]interface{} {
	result := b.Base.Serialize()
	result["Active"] = b.Active
	return result
}
