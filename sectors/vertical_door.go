package sectors

import (
	"tlyakhov/gofoom/core"
	"tlyakhov/gofoom/registry"
)

type DoorBehavior int

//go:generate go run github.com/dmarkham/enumer -type=DoorBehavior -json
const (
	Open DoorBehavior = iota
	Opening
	Closing
	Closed
)

type VerticalDoor struct {
	core.PhysicalSector `editable:"^"`

	OrigTopZ float64
	VelZ     float64
	State    DoorBehavior
}

func init() {
	registry.Instance().Register(VerticalDoor{})
}

func (s *VerticalDoor) Initialize() {
	s.PhysicalSector.Initialize()
	s.OrigTopZ = s.TopZ
}

func (s *VerticalDoor) Deserialize(data map[string]interface{}) {
	s.PhysicalSector.Deserialize(data)
	if v, ok := data["OrigTopZ"]; ok {
		s.OrigTopZ = v.(float64)
	} else {
		s.OrigTopZ = s.TopZ
	}
}

func (s *VerticalDoor) Serialize() map[string]interface{} {
	result := s.PhysicalSector.Serialize()
	result["Type"] = "sectors.VerticalDoor"
	result["OrigTopZ"] = s.OrigTopZ
	return result
}
