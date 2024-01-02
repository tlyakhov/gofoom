package behaviors

import (
	"tlyakhov/gofoom/concepts"
)

type Animated struct {
	concepts.Attached `editable:"^"`

	Animations []concepts.IAnimation `editable:"Animations"`
}

var AnimatedComponentIndex int

func init() {
	AnimatedComponentIndex = concepts.DbTypes().Register(Animated{}, AnimatedFromDb)
}

func AnimatedFromDb(entity *concepts.EntityRef) *Animated {
	if asserted, ok := entity.Component(AnimatedComponentIndex).(*Animated); ok {
		return asserted
	}
	return nil
}

func (w *Animated) String() string {
	return "Animated"
}

func (a *Animated) Construct(data map[string]any) {
	a.Attached.Construct(data)

	a.Animations = make([]concepts.IAnimation, 0)

	if data == nil {
		return
	}
}

func (w *Animated) Serialize() map[string]any {
	result := w.Attached.Serialize()
	return result
}
