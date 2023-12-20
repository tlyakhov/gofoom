package concepts

import (
	"math"
)

type TweeningFunc func(start, end, t float64) float64

var TweeningFuncs = map[string]TweeningFunc{
	"Lerp":         Lerp,
	"EaseIn":       EaseIn,
	"EaseIn3":      EaseIn3,
	"EaseIn4":      EaseIn4,
	"EaseOut":      EaseOut,
	"EaseOut3":     EaseOut3,
	"EaseOut4":     EaseOut4,
	"EaseInOut":    EaseInOut,
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

func Lerp(start, end, t float64) float64 {
	if t <= 0 {
		return start
	}
	if t >= 1 {
		return end
	}
	return start*(1.0-t) + end*t
}

func EaseIn(start, end, t float64) float64 {
	return Lerp(start, end, t*t)
}

func EaseIn3(start, end, t float64) float64 {
	return Lerp(start, end, t*t*t)
}

func EaseIn4(start, end, t float64) float64 {
	return Lerp(start, end, t*t*t*t)
}

func EaseOut(start, end, t float64) float64 {
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

func EaseInOut(start, end, t float64) float64 {
	return Lerp(start, end, Lerp(EaseIn(0, 1, t), EaseOut(0, 1, t), t))
}

func EaseInOut3(start, end, t float64) float64 {
	return Lerp(start, end, Lerp(EaseIn3(0, 1, t), EaseOut3(0, 1, t), t))
}

func EaseInOut4(start, end, t float64) float64 {
	return Lerp(start, end, Lerp(EaseIn4(0, 1, t), EaseOut4(0, 1, t), t))
}

func Spike(start, end, t float64) float64 {
	if t <= 0.5 {
		return Lerp(start, end, t*2.0)
	}
	return Lerp(start, end, (1.0-t)*2.0)
}

func Spike2(start, end, t float64) float64 {
	if t <= 0.5 {
		return EaseIn(start, end, t*2.0)
	}
	return EaseIn(start, end, (1.0-t)*2.0)
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
	p := 0.3
	a := end - start
	s := p / 4
	return -(a * math.Pow(2, 10*(t-1)) * math.Sin((t-s)*(2*math.Pi)/p)) + start
}

func ElasticOut(start, end, t float64) float64 {
	p := 0.3
	a := end - start
	s := p / 4
	return (a * math.Pow(2, -10*t) * math.Sin((t-s)*(2*math.Pi)/p)) + end
}

func ElasticInOut(start, end, t float64) float64 {
	return Lerp(start, end, Lerp(ElasticIn(0, 1, t), ElasticOut(0, 1, t), t))
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
