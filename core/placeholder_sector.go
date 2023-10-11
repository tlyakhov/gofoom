package core

import (
	"tlyakhov/gofoom/concepts"
)

type PlaceholderSector struct {
	concepts.Base
}

func (s *PlaceholderSector) Physical() *PhysicalSector {
	return nil
}

func (s *PlaceholderSector) IsPointInside2D(*concepts.Vector2) bool {
	return false
}
