package core

import "github.com/tlyakhov/gofoom/concepts"

type AbstractSector interface {
	concepts.ISerializable
	IsPointInside2D(p *concepts.Vector2) bool
	Physical() *PhysicalSector
}
