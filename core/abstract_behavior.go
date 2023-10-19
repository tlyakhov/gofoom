package core

import (
	"tlyakhov/gofoom/concepts"
)

type AbstractBehavior interface {
	concepts.ISerializable
	Frame()
	Animated() *AnimatedBehavior
}
