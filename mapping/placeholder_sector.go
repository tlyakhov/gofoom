package mapping

import (
	"github.com/tlyakhov/gofoom/concepts"
)

type PlaceholderSector struct {
	concepts.Base
}

func (s *PlaceholderSector) GetPhysical() *PhysicalSector {
	return nil
}

func (s *PlaceholderSector) IsPointInside2D(*concepts.Vector2) bool {
	return false
}
