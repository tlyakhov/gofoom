package sectors

import (
	"tlyakhov/gofoom/concepts"
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
	concepts.Attached `editable:"^"`
	VelZ              float64
	State             DoorBehavior
}

var VerticalDoorComponentIndex int

func init() {
	VerticalDoorComponentIndex = concepts.DbTypes().Register(VerticalDoor{})
}

func VerticalDoorFromDb(entity *concepts.EntityRef) *VerticalDoor {
	return entity.Component(VerticalDoorComponentIndex).(*VerticalDoor)
}
