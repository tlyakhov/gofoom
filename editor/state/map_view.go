package state

import "github.com/tlyakhov/gofoom/concepts"

type MapView struct {
	Scale        float64
	Pos          concepts.Vector2 // World
	Size         concepts.Vector2 // Screen
	Step         float64          // Grid step
	GridA, GridB concepts.Vector2 // World, lock grid to axis.
}
