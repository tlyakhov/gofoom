package concepts

import (
	"fmt"
	"image/color"
	"math"
)

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

// IntClamp clamps a value between a minimum and maximum.
func IntClamp(x, min, max int) int {
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

func ParseHexColor(hex string) (color.NRGBA, error) {
	var r, g, b, factor uint8
	var n int
	var err error
	if len(hex) == 4 {
		n, err = fmt.Sscanf(hex, "#%1x%1x%1x", &r, &g, &b)
		factor = 16
	} else {
		n, err = fmt.Sscanf(hex, "#%2x%2x%2x", &r, &g, &b)
		factor = 1
	}
	if err != nil {
		return color.NRGBA{}, err
	}
	if n != 3 {
		return color.NRGBA{}, fmt.Errorf("color %v is not a hex-color", hex)
	}
	return color.NRGBA{r * factor, g * factor, b * factor, 255}, nil
}
