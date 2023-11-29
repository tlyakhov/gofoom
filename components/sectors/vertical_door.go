package sectors

import (
	"tlyakhov/gofoom/concepts"
)

type DoorState int

//go:generate go run github.com/dmarkham/enumer -type=DoorState -json
const (
	DoorStateOpen DoorState = iota
	DoorStateOpening
	DoorStateClosing
	DoorStateClosed
)

type DoorIntent int

//go:generate go run github.com/dmarkham/enumer -type=DoorIntent -json
const (
	DoorIntentReset DoorIntent = iota
	DoorIntentOpen
	DoorIntentClosed
)

type VerticalDoor struct {
	concepts.Attached `editable:"^"`
	VelZ              float64
	State             DoorState
	Intent            DoorIntent
}

var VerticalDoorComponentIndex int

func init() {
	VerticalDoorComponentIndex = concepts.DbTypes().Register(VerticalDoor{}, VerticalDoorFromDb)
	dis := DoorIntentStrings()
	div := DoorIntentValues()
	for i, s := range dis {
		concepts.DbTypes().ExprEnv[s] = div[i]
	}
}

func VerticalDoorFromDb(entity *concepts.EntityRef) *VerticalDoor {
	if asserted, ok := entity.Component(VerticalDoorComponentIndex).(*VerticalDoor); ok {
		return asserted
	}
	return nil
}

func (vd *VerticalDoor) String() string {
	return "VerticalDoor"
}
