// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"reflect"
	"tlyakhov/gofoom/components/core"
	"tlyakhov/gofoom/dynamic"
	"tlyakhov/gofoom/ecs"

	"github.com/spf13/cast"
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
	// TODO: Separate out state and make this multi-attachable
	State     DoorState
	Intent    DoorIntent  `editable:"Intent"`
	AutoClose bool        `editable:"Auto-close"`
	Open      core.Script `editable:"On Open"`
	Close     core.Script `editable:"On Close"`

	// Passthroughs to Animation
	TweeningFunc dynamic.TweeningFunc `editable:"Tweening Function"`
	Duration     float64              `editable:"Duration"`
}

func init() {
	dis := DoorIntentStrings()
	div := DoorIntentValues()
	for i, s := range dis {
		ecs.Types().ExprEnv[s] = div[i]
	}
}

func (vd *VerticalDoor) String() string {
	return "VerticalDoor"
}

func (vd *VerticalDoor) Construct(data map[string]any) {
	vd.Attached.Construct(data)
	vd.AutoClose = true
	vd.TweeningFunc = dynamic.EaseInOut2
	vd.Duration = 1000

	if data == nil {
		vd.Open.Construct(nil)
		vd.Close.Construct(nil)
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
		vd.AutoClose = cast.ToBool(v)
	}
	if v, ok := data["TweeningFunc"]; ok {
		name := v.(string)
		vd.TweeningFunc = dynamic.TweeningFuncs[name]
		if vd.TweeningFunc == nil {
			vd.TweeningFunc = dynamic.EaseInOut2
		}
	}
	if v, ok := data["Duration"]; ok {
		vd.Duration = cast.ToFloat64(v)
	}
	if v, ok := data["Open"]; ok {
		vd.Open.Construct(v.(map[string]any))
	} else {
		vd.Open.Construct(nil)
	}
	if v, ok := data["Close"]; ok {
		vd.Close.Construct(v.(map[string]any))
	} else {
		vd.Close.Construct(nil)
	}
}

func (vd *VerticalDoor) Serialize() map[string]any {
	result := vd.Attached.Serialize()

	if !vd.AutoClose {
		result["AutoClose"] = false
	}
	if !vd.Open.IsEmpty() {
		result["Open"] = vd.Open.Serialize()
	}
	if !vd.Close.IsEmpty() {
		result["Close"] = vd.Close.Serialize()
	}
	result["Duration"] = vd.Duration
	result["TweeningFunc"] = dynamic.TweeningFuncNames[reflect.ValueOf(vd.TweeningFunc).Pointer()]
	return result
}
