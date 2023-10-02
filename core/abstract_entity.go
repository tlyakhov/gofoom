package core

import "tlyakhov/gofoom/concepts"

type AbstractEntity interface {
	concepts.ISerializable
	GetSector() AbstractSector
	Physical() *PhysicalEntity
}
