package core

import "tlyakhov/gofoom/concepts"

type AbstractMob interface {
	concepts.ISerializable
	GetSector() AbstractSector
	Physical() *PhysicalMob
}
