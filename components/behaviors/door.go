// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package behaviors

import (
	"log"
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

type DoorType int

//go:generate go run github.com/dmarkham/enumer -type=DoorType -json
const (
	DoorTypeVertical DoorType = iota
	DoorTypeSwing
)

type Door struct {
	ecs.Attached `editable:"^"`
	// TODO: Separate out state and make this shareable
	State         DoorState
	Intent        DoorIntent  `editable:"Intent"`
	Type          DoorType    `editable:"Type"`
	AutoClose     bool        `editable:"Auto-close"`
	Open          core.Script `editable:"On Open"`
	Close         core.Script `editable:"On Close"`
	AutoProximity bool        `editable:"Auto Proximity"`
	Limit         float64     `editable:"Limit"`
	UseLimit      bool        `editable:"Use Limit?"`

	// Passthroughs to Animation
	TweeningFunc dynamic.TweeningFunc `editable:"Tweening Function"`
	Duration     float64              `editable:"Duration"`
}

func init() {
	// TODO: Generalize this to all enums, potentially include in code generator
	dis := DoorIntentStrings()
	div := DoorIntentValues()
	for i, s := range dis {
		ecs.Types().ExprEnv[s] = div[i]
	}
}

func (d *Door) String() string {
	return "Door"
}

func (d *Door) Construct(data map[string]any) {
	d.Attached.Construct(data)
	d.AutoClose = true
	d.AutoProximity = true
	d.TweeningFunc = dynamic.EaseInOut2
	d.Duration = 1000
	d.Type = DoorTypeVertical
	d.UseLimit = false

	if data == nil {
		d.Open.Construct(nil)
		d.Close.Construct(nil)
		return
	}

	if v, ok := data["Intent"]; ok {
		if intent, err := DoorIntentString(v.(string)); err == nil {
			d.Intent = intent
		} else {
			log.Printf("error parsing door intent %v", v)
		}
	}

	if v, ok := data["Type"]; ok {
		if t, err := DoorTypeString(v.(string)); err == nil {
			d.Type = t
		} else {
			log.Printf("error parsing door type %v", v)
		}
	}

	if v, ok := data["AutoClose"]; ok {
		d.AutoClose = cast.ToBool(v)
	}
	if v, ok := data["AutoProximity"]; ok {
		d.AutoProximity = cast.ToBool(v)
	}
	if v, ok := data["TweeningFunc"]; ok {
		name := v.(string)
		d.TweeningFunc = dynamic.TweeningFuncs[name]
		if d.TweeningFunc == nil {
			d.TweeningFunc = dynamic.EaseInOut2
		}
	}
	if v, ok := data["Duration"]; ok {
		d.Duration = cast.ToFloat64(v)
	}
	if v, ok := data["Limit"]; ok {
		d.Limit = cast.ToFloat64(v)
	}
	if v, ok := data["UseLimit"]; ok {
		d.UseLimit = cast.ToBool(v)
	}
	if v, ok := data["Open"]; ok {
		d.Open.Construct(v.(map[string]any))
	} else {
		d.Open.Construct(nil)
	}
	if v, ok := data["Close"]; ok {
		d.Close.Construct(v.(map[string]any))
	} else {
		d.Close.Construct(nil)
	}
}

func (d *Door) Serialize() map[string]any {
	result := d.Attached.Serialize()

	if !d.AutoClose {
		result["AutoClose"] = false
	}
	if !d.AutoProximity {
		result["AutoProximity"] = false
	}
	if !d.Open.IsEmpty() {
		result["Open"] = d.Open.Serialize()
	}
	if !d.Close.IsEmpty() {
		result["Close"] = d.Close.Serialize()
	}
	result["Intent"] = d.Intent.String()
	result["Type"] = d.Type.String()
	result["Duration"] = d.Duration
	result["Limit"] = d.Limit
	result["UseLimit"] = d.UseLimit
	result["TweeningFunc"] = dynamic.TweeningFuncNames[reflect.ValueOf(d.TweeningFunc).Pointer()]
	return result
}
