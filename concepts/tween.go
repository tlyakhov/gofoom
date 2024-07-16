// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"math"
	"reflect"
)

type TweeningFunc func(start, end, t float64) float64

var TweeningFuncs = map[string]TweeningFunc{
	"Lerp":         Lerp,
	"EaseIn2":      EaseIn2,
	"EaseIn3":      EaseIn3,
	"EaseIn4":      EaseIn4,
	"EaseOut2":     EaseOut2,
	"EaseOut3":     EaseOut3,
	"EaseOut4":     EaseOut4,
	"EaseInOut2":   EaseInOut2,
	"EaseInOut3":   EaseInOut3,
	"EaseInOut4":   EaseInOut4,
	"Spike":        Spike,
	"Spike2":       Spike2,
	"Spike3":       Spike3,
	"Spike4":       Spike4,
	"ElasticIn":    ElasticIn,
	"ElasticOut":   ElasticOut,
	"ElasticInOut": ElasticInOut,
}
var TweeningFuncNames map[uintptr]string

func init() {
	TweeningFuncNames = make(map[uintptr]string)
	for name, f := range TweeningFuncs {
		TweeningFuncNames[reflect.ValueOf(f).Pointer()] = name
	}
}

func Lerp(start, end, t float64) float64 {
	if t <= 0 {
		return start
	}
	if t >= 1 {
		return end
	}
	return start*(1.0-t) + end*t
}

func EaseIn2(start, end, t float64) float64 {
	return Lerp(start, end, t*t)
}

func EaseIn3(start, end, t float64) float64 {
	return Lerp(start, end, t*t*t)
}

func EaseIn4(start, end, t float64) float64 {
	return Lerp(start, end, t*t*t*t)
}

func EaseOut2(start, end, t float64) float64 {
	inv := 1.0 - t
	return Lerp(start, end, 1.0-inv*inv)
}

func EaseOut3(start, end, t float64) float64 {
	inv := 1.0 - t
	return Lerp(start, end, 1.0-inv*inv*inv)
}

func EaseOut4(start, end, t float64) float64 {
	inv := 1.0 - t
	return Lerp(start, end, 1.0-inv*inv*inv*inv)
}

func EaseInOut2(start, end, t float64) float64 {
	if t < 0.5 {
		t = 2 * t * t
	} else {
		t = (-2*t + 2)
		t *= t
		t = 1 - t*0.5
	}
	return Lerp(start, end, t)
}

func EaseInOut3(start, end, t float64) float64 {
	if t < 0.5 {
		t = 4 * t * t
	} else {
		t = (-2*t + 2)
		t = t * t * t
		t = 1 - t*0.5
	}
	return Lerp(start, end, t)
}

func EaseInOut4(start, end, t float64) float64 {
	if t < 0.5 {
		t = 8 * t * t * t * t
	} else {
		t = (-2*t + 2)
		t = t * t * t * t
		t = 1 - t*0.5
	}
	return Lerp(start, end, t)
}

func Spike(start, end, t float64) float64 {
	if t <= 0.5 {
		return Lerp(start, end, t*2.0)
	}
	return Lerp(start, end, (1.0-t)*2.0)
}

func Spike2(start, end, t float64) float64 {
	if t <= 0.5 {
		return EaseIn2(start, end, t*2.0)
	}
	return EaseIn2(start, end, (1.0-t)*2.0)
}

func Spike3(start, end, t float64) float64 {
	if t <= 0.5 {
		return EaseIn3(start, end, t*2.0)
	}
	return EaseIn3(start, end, (1.0-t)*2.0)
}

func Spike4(start, end, t float64) float64 {
	if t <= 0.5 {
		return EaseIn4(start, end, t*2.0)
	}
	return EaseIn4(start, end, (1.0-t)*2.0)
}

func ElasticIn(start, end, t float64) float64 {
	const c4 = (2 * math.Pi) / 3
	if t == 0 {
		return start
	}
	if t == 1 {
		return end
	}

	t = -math.Pow(2, 10*t-10) * math.Sin((t*10-10.75)*c4)
	return start*(1.0-t) + end*t
}

func ElasticOut(start, end, t float64) float64 {
	const c4 = (2 * math.Pi) / 3
	if t == 0 {
		return start
	}
	if t == 1 {
		return end
	}
	t = math.Pow(2, -10*t)*math.Sin((t*10-0.75)*c4) + 1
	return start*(1.0-t) + end*t
}

func ElasticInOut(start, end, t float64) float64 {
	const c5 = (2 * math.Pi) / 4.5
	if t == 0 {
		return start
	}
	if t == 1 {
		return end
	}
	if t < 0.5 {
		t = -(math.Pow(2, 20*t-10) * math.Sin((20*t-11.125)*c5)) / 2
	} else {
		t = (math.Pow(2, -20*t+10)*math.Sin((20*t-11.125)*c5))/2 + 1
	}
	return start*(1.0-t) + end*t
}

func TweenAngles(start, end, t float64, f TweeningFunc) float64 {
	// Simple way to do this is
	// to convert to cartesian space, blend, convert back
	y1, x1 := math.Sincos(start * Deg2rad)
	y2, x2 := math.Sincos(end * Deg2rad)
	x := f(x1, x2, t)
	y := f(y1, y2, t)
	return math.Atan2(y, x) * Rad2deg
}
