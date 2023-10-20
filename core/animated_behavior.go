package core

import (
	"tlyakhov/gofoom/concepts"
)

type AnimatedBehavior struct {
	concepts.Base `editable:"^"`
	Entity        AbstractEntity

	Active bool `editable:"Active?"`
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
