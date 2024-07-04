// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"reflect"
	"tlyakhov/gofoom/constants"
)

//go:generate go run github.com/dmarkham/enumer -type=AnimationLifetime -json
type AnimationLifetime int

const (
	AnimationLifetimeOnce AnimationLifetime = iota
	AnimationLifetimeHold
	AnimationLifetimeLoop
)

//go:generate go run github.com/dmarkham/enumer -type=AnimationCoordinates -json
type AnimationCoordinates int

const (
	AnimationCoordinatesRelative AnimationCoordinates = iota
	AnimationCoordinatesAbsolute
)

type Animated interface {
	Serializable
	Animate()
	Reset()
}

type Animation[T Simulatable] struct {
	*SimVariable[T]
	TweeningFunc `editable:"Tweening Function"`

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
	if a == nil || !a.Active || a.SimVariable == nil {
		return
	}
	a.Percent = Clamp(a.Percent, 0, 1)
	switch c := any(a).(type) {
	case *Animation[int]:
		c.Now = int(c.TweeningFunc(float64(c.Start), float64(c.End), c.Percent))
		if a.Coordinates == AnimationCoordinatesRelative {
			c.Now += c.Original
		}
	case *Animation[float64]:
		c.Now = c.TweeningFunc(c.Start, c.End, c.Percent)
		if a.Coordinates == AnimationCoordinatesRelative {
			c.Now += c.Original
		}
	case *Animation[Vector2]:
		c.Now[0] = c.TweeningFunc(c.Start[0], c.End[0], c.Percent)
		c.Now[1] = c.TweeningFunc(c.Start[1], c.End[1], c.Percent)
		if a.Coordinates == AnimationCoordinatesRelative {
			c.Now.AddSelf(&c.Original)
		}
	case *Animation[Vector3]:
		c.Now[0] = c.TweeningFunc(c.Start[0], c.End[0], c.Percent)
		c.Now[1] = c.TweeningFunc(c.Start[1], c.End[1], c.Percent)
		c.Now[2] = c.TweeningFunc(c.Start[2], c.End[2], c.Percent)
		if a.Coordinates == AnimationCoordinatesRelative {
			c.Now.AddSelf(&c.Original)
		}
	case *Animation[Vector4]:
		c.Now[0] = c.TweeningFunc(c.Start[0], c.End[0], c.Percent)
		c.Now[1] = c.TweeningFunc(c.Start[1], c.End[1], c.Percent)
		c.Now[2] = c.TweeningFunc(c.Start[2], c.End[2], c.Percent)
		c.Now[3] = c.TweeningFunc(c.Start[3], c.End[3], c.Percent)
		if a.Coordinates == AnimationCoordinatesRelative {
			c.Now.AddSelf(&c.Original)
		}
	}

	if (a.Percent >= 1 && !a.Reverse) || (a.Percent <= 0 && a.Reverse) {
		switch a.Lifetime {
		case AnimationLifetimeOnce:
			a.Active = false
		case AnimationLifetimeLoop:
			a.Reverse = !a.Reverse
		}
	}
	if a.Reverse {
		a.Percent -= constants.TimeStep / a.Duration
	} else {
		a.Percent += constants.TimeStep / a.Duration
	}
}

func (a *Animation[T]) Construct(data map[string]any) {
	a.Active = true
	a.Duration = 1000
	a.Reverse = false
	a.TweeningFunc = Lerp
	a.Lifetime = AnimationLifetimeLoop
	a.Coordinates = AnimationCoordinatesRelative

	if data == nil {
		return
	}

	if v, ok := data["TweeningFunc"]; ok {
		name := v.(string)
		a.TweeningFunc = TweeningFuncs[name]
	}
	if v, ok := data["Duration"]; ok {
		a.Duration = v.(float64)
	}
	if v, ok := data["Percent"]; ok {
		a.Percent = v.(float64)
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
			panic(err)
		}
	}
	if v, ok := data["Coordinates"]; ok {
		acs, err := AnimationCoordinatesString(v.(string))
		if err == nil {
			a.Coordinates = acs
		} else {
			panic(err)
		}
	}

	switch c := any(a).(type) {
	case *Animation[int]:
		if v, ok := data["Start"]; ok {
			c.Start = v.(int)
		}
		if v, ok := data["End"]; ok {
			c.End = v.(int)
		}
	case *Animation[float64]:
		if v, ok := data["Start"]; ok {
			c.Start = v.(float64)
		}
		if v, ok := data["End"]; ok {
			c.End = v.(float64)
		}
	case *Animation[Vector2]:
		if v, ok := data["Start"]; ok {
			c.Start.Deserialize(v.(map[string]any))
		}
		if v, ok := data["End"]; ok {
			c.End.Deserialize(v.(map[string]any))
		}
	case *Animation[Vector3]:
		if v, ok := data["Start"]; ok {
			c.Start.Deserialize(v.(map[string]any))
		}
		if v, ok := data["End"]; ok {
			c.End.Deserialize(v.(map[string]any))
		}
	case *Animation[Vector4]:
		if v, ok := data["Start"]; ok {
			c.Start.Deserialize(v.(map[string]any), false)
		}
		if v, ok := data["End"]; ok {
			c.End.Deserialize(v.(map[string]any), false)
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
	if a.Lifetime != AnimationLifetimeLoop {
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
	case *Animation[Vector2]:
		result["Start"] = c.Start.Serialize()
		result["End"] = c.End.Serialize()
	case *Animation[Vector3]:
		result["Start"] = c.Start.Serialize()
		result["End"] = c.End.Serialize()
	case *Animation[Vector4]:
		result["Start"] = c.Start.Serialize(false)
		result["End"] = c.End.Serialize(false)
	}
	return result
}

func (a *Animation[T]) SetDB(db *EntityComponentDB) {
}
func (a *Animation[T]) GetDB() *EntityComponentDB {
	return nil
}
