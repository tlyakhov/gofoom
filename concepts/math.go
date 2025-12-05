// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"cmp"
	"math"
)

const (
	Deg2rad float64 = math.Pi / 180.0
	Rad2deg float64 = 180.0 / math.Pi
)

// Clamp clamps a value between a minimum and maximum.
func Clamp[T cmp.Ordered](x, min, max T) T {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

func NearestPow2(n uint32) uint32 {
	if n == 0 {
		return 0
	}
	n |= (n >> 1)
	n |= (n >> 2)
	n |= (n >> 4)
	n |= (n >> 8)
	n |= (n >> 16)
	return n - (n >> 1)
}

func ByteClamp(x float64) uint8 {
	if x <= 0 {
		return 0
	}
	if x >= 0xFF {
		return 0xFF
	}

	return uint8(x)
}

func NormalizeAngle(a float64) float64 {
	a = math.Mod(a, 360) + 360
	if a < 0 {
		return a + 360
	}
	return a
}

func MinimizeAngleDistance(src float64, dst *float64) {
	for *dst-src > 180 {
		*dst -= 360
	}
	for *dst-src < -180 {
		*dst += 360
	}
}

func NanosToMillis(nanos int64) float64 {
	return float64(nanos) / 1_000_000
}

func MillisToNanos(millis float64) int64 {
	return int64(millis * 1_000_000)
}
