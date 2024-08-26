// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"reflect"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/ecs"
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
	ecs.Attached `editable:"^"`
	State        DoorState
	Intent       DoorIntent `editable:"Intent"`
	AutoClose    bool       `editable:"Auto-close"`

	// Passthroughs to Animation
	TweeningFunc concepts.TweeningFunc `editable:"Tweening Function"`
	Duration     float64               `editable:"Duration"`
}

var VerticalDoorComponentIndex int

func init() {
	VerticalDoorComponentIndex = ecs.RegisterComponent(&ecs.Column[VerticalDoor, *VerticalDoor]{Getter: GetVerticalDoor})
	dis := DoorIntentStrings()
	div := DoorIntentValues()
	for i, s := range dis {
		ecs.Types().ExprEnv[s] = div[i]
	}
}

func GetVerticalDoor(db *ecs.ECS, e ecs.Entity) *VerticalDoor {
	if asserted, ok := db.Component(e, VerticalDoorComponentIndex).(*VerticalDoor); ok {
		return asserted
	}
	return nil
}

func (vd *VerticalDoor) String() string {
	return "VerticalDoor"
}

func (vd *VerticalDoor) Construct(data map[string]any) {
	vd.Attached.Construct(data)
	vd.AutoClose = true
	vd.TweeningFunc = concepts.EaseInOut2
	vd.Duration = 1000

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

	if v, ok := data["AutoClose"]; ok {
		vd.AutoClose = v.(bool)
	}
	if v, ok := data["TweeningFunc"]; ok {
		name := v.(string)
		vd.TweeningFunc = concepts.TweeningFuncs[name]
		if vd.TweeningFunc == nil {
			vd.TweeningFunc = concepts.EaseInOut2
		}
	}
	if v, ok := data["Duration"]; ok {
		vd.Duration = v.(float64)
	}
}

func (vd *VerticalDoor) Serialize() map[string]any {
	result := vd.Attached.Serialize()

	if !vd.AutoClose {
		result["AutoClose"] = false
	}
	result["Duration"] = vd.Duration
	result["TweeningFunc"] = concepts.TweeningFuncNames[reflect.ValueOf(vd.TweeningFunc).Pointer()]
	return result
}
