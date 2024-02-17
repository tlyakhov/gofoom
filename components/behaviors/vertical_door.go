package behaviors

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

func (vd *VerticalDoor) Construct(data map[string]any) {
	vd.Attached.Construct(data)

	if data == nil {
		return
	}

	if v, ok := data["Intent"]; ok {
		if intent, err := DoorIntentString(v.(string)); err == nil {
			vd.Intent = intent
		} else {
			panic(err)
		}
	}
}
