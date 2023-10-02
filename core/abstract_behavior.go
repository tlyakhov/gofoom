package core

import (
	"tlyakhov/gofoom/concepts"
)

type AbstractBehavior interface {
	concepts.ISerializable
	Frame(lastFrameTime float64)
	Animated() *AnimatedBehavior
}
