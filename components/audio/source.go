package audio

import (
	"tlyakhov/gofoom/ecs"
)

type Source struct {
	ecs.Attached
}

func (src *Source) String() string {
	return "Audio Source"
}

func (src *Source) Construct(data map[string]any) {
	src.Attached.Construct(data)

	if data == nil {
		return
	}
}

func (src *Source) Serialize() map[string]any {
	result := src.Attached.Serialize()

	return result
}
