package archetypes

import "tlyakhov/gofoom/concepts"

type Archetype interface {
	Is(entity concepts.EntityRef) bool
	Create() concepts.EntityRef
}
