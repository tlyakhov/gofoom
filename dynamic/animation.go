// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package dynamic

import (
	"log"
	"reflect"
	"tlyakhov/gofoom/concepts"
	"tlyakhov/gofoom/constants"

	"github.com/spf13/cast"
)

//go:generate go run github.com/dmarkham/enumer -type=AnimationLifetime -json
type AnimationLifetime int

const (
	AnimationLifetimeOnce AnimationLifetime = iota
	AnimationLifetimeLoop
	AnimationLifetimeBounceOnce
	AnimationLifetimeBounce
)

//go:generate go run github.com/dmarkham/enumer -type=AnimationCoordinates -json
type AnimationCoordinates int

const (
	AnimationCoordinatesRelative AnimationCoordinates = iota
	AnimationCoordinatesAbsolute
)

type Animated interface {
	Animate()
	Reset()
}

type Animation[T DynamicType] struct {
	*DynamicValue[T]
	TweeningFunc `editable:"Function"`

	Start       T                    `editable:"Start"`
	End         T                    `editable:"End"`
	Duration    float64              `editable:"Duration"`
	Active      bool                 `editable:"Active"`
	Reverse     bool                 `editable:"Reverse"`
	Lifetime    AnimationLifetime    `editable:"Lifetime"`
	Coordinates AnimationCoordinates `editable:"Coordinates"`
	Percent     float64
}

func (a *Animation[T]) SetEasingFunc(name string) TweeningFunc {
	a.TweeningFunc = TweeningFuncs[name]
	if a.TweeningFunc == nil {
		a.TweeningFunc = Lerp
	}
	return a.TweeningFunc
}

func (a *Animation[T]) Reset() {
	if a == nil {
		return
	}
	if a.Reverse {
		a.Percent = 1
	} else {
		a.Percent = 0
	}
}

func (a *Animation[T]) Animate() {
	if a == nil || !a.Active || a.DynamicValue == nil {
		return
	}
	if a.Reverse {
		a.Percent -= constants.TimeStep / a.Duration
	} else {
		a.Percent += constants.TimeStep / a.Duration
	}
	a.Percent = concepts.Clamp(a.Percent, 0, 1)
	percent := a.Percent
	if a.Lifetime == AnimationLifetimeBounce || a.Lifetime == AnimationLifetimeBounceOnce {
		percent *= 2
		if percent > 1 {
			percent = 2.0 - percent
		}
	}
	switch c := any(a).(type) {
	case *Animation[int]:
		c.Now = int(c.TweeningFunc(float64(c.Start), float64(c.End), percent))
		if a.Coordinates == AnimationCoordinatesRelative {
			c.Now += c.Spawn
		}
	case *Animation[float64]:
		c.Now = c.TweeningFunc(c.Start, c.End, percent)
		if a.Coordinates == AnimationCoordinatesRelative {
			c.Now += c.Spawn
		}
	case *Animation[concepts.Vector2]:
		c.Now[0] = c.TweeningFunc(c.Start[0], c.End[0], percent)
		c.Now[1] = c.TweeningFunc(c.Start[1], c.End[1], percent)
		if a.Coordinates == AnimationCoordinatesRelative {
			c.Now.AddSelf(&c.Spawn)
		}
	case *Animation[concepts.Vector3]:
		c.Now[0] = c.TweeningFunc(c.Start[0], c.End[0], percent)
		c.Now[1] = c.TweeningFunc(c.Start[1], c.End[1], percent)
		c.Now[2] = c.TweeningFunc(c.Start[2], c.End[2], percent)
		if a.Coordinates == AnimationCoordinatesRelative {
			c.Now.AddSelf(&c.Spawn)
		}
	case *Animation[concepts.Vector4]:
		c.Now[0] = c.TweeningFunc(c.Start[0], c.End[0], percent)
		c.Now[1] = c.TweeningFunc(c.Start[1], c.End[1], percent)
		c.Now[2] = c.TweeningFunc(c.Start[2], c.End[2], percent)
		c.Now[3] = c.TweeningFunc(c.Start[3], c.End[3], percent)
		if a.Coordinates == AnimationCoordinatesRelative {
			c.Now.AddSelf(&c.Spawn)
		}
	}

	if (a.Percent >= 1 && !a.Reverse) || (a.Percent <= 0 && a.Reverse) {
		switch a.Lifetime {
		case AnimationLifetimeOnce:
			fallthrough
		case AnimationLifetimeBounceOnce:
			a.Active = false
		case AnimationLifetimeBounce:
			a.Reverse = !a.Reverse
		case AnimationLifetimeLoop:
			a.Reset()
		}
	}
}

func (a *Animation[T]) Construct(data map[string]any) {
	a.Active = true
	a.Duration = 1000
	a.Reverse = false
	a.TweeningFunc = Lerp
	a.Lifetime = AnimationLifetimeBounce
	a.Coordinates = AnimationCoordinatesRelative

	if data == nil {
		return
	}

	if v, ok := data["TweeningFunc"]; ok {
		name := v.(string)
		a.TweeningFunc = TweeningFuncs[name]
		if a.TweeningFunc == nil {
			a.TweeningFunc = Lerp
		}
	}
	if v, ok := data["Duration"]; ok {
		a.Duration = cast.ToFloat64(v)
	}
	if v, ok := data["Percent"]; ok {
		a.Percent = cast.ToFloat64(v)
	}
	if v, ok := data["Active"]; ok {
		a.Active = v.(bool)
	}
	if v, ok := data["Reverse"]; ok {
		a.Reverse = v.(bool)
	}
	if v, ok := data["Lifetime"]; ok {
		als, err := AnimationLifetimeString(v.(string))
		if err == nil {
			a.Lifetime = als
		} else {
			log.Printf("Animation.Construct: %v", err)
		}
	}
	if v, ok := data["Coordinates"]; ok {
		acs, err := AnimationCoordinatesString(v.(string))
		if err == nil {
			a.Coordinates = acs
		} else {
			log.Printf("Animation.Construct: %v", err)
		}
	}

	switch c := any(a).(type) {
	case *Animation[int]:
		if v, ok := data["Start"]; ok {
			c.Start = cast.ToInt(v)
		}
		if v, ok := data["End"]; ok {
			c.End = cast.ToInt(v)
		}
	case *Animation[float64]:
		if v, ok := data["Start"]; ok {
			c.Start = cast.ToFloat64(v)
		}
		if v, ok := data["End"]; ok {
			c.End = cast.ToFloat64(v)
		}
	case *Animation[concepts.Vector2]:
		if v, ok := data["Start"]; ok {
			c.Start.Deserialize(v.(string))
		}
		if v, ok := data["End"]; ok {
			c.End.Deserialize(v.(string))
		}
	case *Animation[concepts.Vector3]:
		if v, ok := data["Start"]; ok {
			c.Start.Deserialize(v.(string))
		}
		if v, ok := data["End"]; ok {
			c.End.Deserialize(v.(string))
		}
	case *Animation[concepts.Vector4]:
		if v, ok := data["Start"]; ok {
			c.Start.Deserialize(v.(string))
		}
		if v, ok := data["End"]; ok {
			c.End.Deserialize(v.(string))
		}
	}
}
func (a *Animation[T]) Serialize() map[string]any {
	result := make(map[string]any)
	result["Duration"] = a.Duration
	result["Active"] = a.Active
	result["Reverse"] = a.Reverse
	result["Percent"] = a.Percent
	result["TweeningFunc"] = TweeningFuncNames[reflect.ValueOf(a.TweeningFunc).Pointer()]
	if a.Lifetime != AnimationLifetimeBounce {
		result["Lifetime"] = a.Lifetime.String()
	}
	if a.Coordinates != AnimationCoordinatesRelative {
		result["Coordinates"] = a.Coordinates.String()
	}

	switch c := any(a).(type) {
	case *Animation[int]:
		result["Start"] = c.Start
		result["End"] = c.End
	case *Animation[float64]:
		result["Start"] = c.Start
		result["End"] = c.End
	case *Animation[concepts.Vector2]:
		result["Start"] = c.Start.Serialize()
		result["End"] = c.End.Serialize()
	case *Animation[concepts.Vector3]:
		result["Start"] = c.Start.Serialize()
		result["End"] = c.End.Serialize()
	case *Animation[concepts.Vector4]:
		result["Start"] = c.Start.Serialize(false)
		result["End"] = c.End.Serialize(false)
	}
	return result
}
