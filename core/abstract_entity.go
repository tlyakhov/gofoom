package core

import "github.com/tlyakhov/gofoom/concepts"

type AbstractEntity interface {
	concepts.ISerializable
	GetSector() AbstractSector
	Physical() *PhysicalEntity
	Behaviors() map[string]AbstractBehavior
}
