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
	VelZ                float64
	State               DoorBehavior
}

func init() {
	registry.Instance().Register(VerticalDoor{})
}
