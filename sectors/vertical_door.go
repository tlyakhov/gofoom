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

	Pos   core.SimScalar
	VelZ  float64
	State DoorBehavior
}

func init() {
	registry.Instance().Register(VerticalDoor{})
}

func (s *VerticalDoor) Apply() {
	s.TopZ = s.Pos.Render
}

func (s *VerticalDoor) Initialize() {
	s.PhysicalSector.Initialize()
	s.Pos.RenderCallback = s.Apply
	s.Pos.Original = s.TopZ
	s.Pos.Reset()
}

func (s *VerticalDoor) Attach(sim *core.Simulation) {
	s.PhysicalSector.Attach(sim)
	sim.AllScalars[&s.Pos] = true
}
func (s *VerticalDoor) Detach() {
	delete(s.Simulation.AllScalars, &s.Pos)
	s.PhysicalSector.Detach()
}

func (s *VerticalDoor) Deserialize(data map[string]interface{}) {
	s.PhysicalSector.Deserialize(data)
	s.Pos.RenderCallback = s.Apply
	if v, ok := data["OrigTopZ"]; ok {
		s.Pos.Original = v.(float64)
	} else {
		s.Pos.Original = s.TopZ
	}
	s.Pos.Reset()
}

func (s *VerticalDoor) Serialize() map[string]interface{} {
	result := s.PhysicalSector.Serialize()
	result["Type"] = "sectors.VerticalDoor"
	result["OrigTopZ"] = s.Pos.Original
	return result
}
