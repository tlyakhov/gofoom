package core

import (
	"github.com/tlyakhov/gofoom/concepts"
)

type AbstractBehavior interface {
	concepts.ISerializable
	Frame(lastFrameTime float64)
	Animated() *AnimatedBehavior
}
