package math

import "math"

const (
	Deg2rad float64 = math.Pi / 180.0
	Rad2deg float64 = 180.0 / math.Pi
)

// Clamp clamps a value between a minimum and maximum.
func Clamp(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

func NearestPow2(n uint) uint {
	n--

	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16

	n++

	return n
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func UMin(x, y uint) uint {
	if x < y {
		return x
	}
	return y
}

func UMax(x, y uint) uint {
	if x > y {
		return x
	}
	return y
}

func NormalizeAngle(a float64) float64 {
	for a < 0 {
		a += 360.0
	}
	for a >= 360.0 {
		a -= 360.0
	}
	return a
}
