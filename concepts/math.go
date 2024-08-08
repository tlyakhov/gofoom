// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"cmp"
	"fmt"
	"image/color"
	"log"
	"math"
	"time"
	"tlyakhov/gofoom/constants"
	"unicode"
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

func Min[T cmp.Ordered](x, y T) T {
	if x < y {
		return x
	}
	return y
}

func Max[T cmp.Ordered](x, y T) T {
	if x > y {
		return x
	}
	return y
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

func ColorToInt32PreMul(c color.Color) uint32 {
	r, g, b, a := c.RGBA()
	r = r >> 8
	g = g >> 8
	b = b >> 8
	a = a >> 8

	return ((r & 0xFF) << 24) | ((g & 0xFF) << 16) | ((b & 0xFF) << 8) | (a & 0xFF)
}

func NRGBAToInt32(c color.NRGBA) uint32 {
	return uint32(c.R)<<24 | uint32(c.G)<<16 | uint32(c.B)<<8 | uint32(c.A)
}
func RGBAToInt32(c color.RGBA) uint32 {
	return uint32(c.R)<<24 | uint32(c.G)<<16 | uint32(c.B)<<8 | uint32(c.A)
}

func Int32ToNRGBA(c uint32) color.NRGBA {
	return color.NRGBA{uint8((c >> 24) & 0xFF), uint8((c >> 16) & 0xFF), uint8((c >> 8) & 0xFF), uint8(c & 0xFF)}
}

func Int32ToRGBA(c uint32) color.RGBA {
	return color.RGBA{uint8((c >> 24) & 0xFF), uint8((c >> 16) & 0xFF), uint8((c >> 8) & 0xFF), uint8(c & 0xFF)}
}

func Int32ToVector3(c uint32) Vector3 {
	return Vector3{float64((c >> 24) & 0xFF), float64((c >> 16) & 0xFF), float64((c >> 8) & 0xFF)}
}

func RngXorShift64(xorSeed uint64) uint64 {
	if xorSeed == 0 {
		xorSeed = uint64(time.Now().UnixNano())
	}
	x := xorSeed * uint64(2685821657736338717)
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	/*x ^= x << 13
	x ^= x >> 7
	x ^= x << 17*/
	return x
}

func TruncateString(str string, max int) string {
	lastSpaceIx := -1
	len := 0
	for i, r := range str {
		if unicode.IsSpace(r) {
			lastSpaceIx = i
		}
		len++
		if len >= max {
			if lastSpaceIx != -1 {
				return str[:lastSpaceIx] + "..."
			}
			// If here, string is longer than max, but has no spaces
			return str[:max] + "..."
		}
	}
	// If here, string is shorter than max
	return str
}

// https://gamedev.stackexchange.com/questions/96459/fast-ray-sphere-collision-code
func IntersectLineSphere(a, b, c *Vector3, r float64) bool {
	dz := b[2] - a[2]
	dy := b[1] - a[1]
	dx := b[0] - a[0]
	t := dx*dx + dy*dy + dz*dz
	dl := math.Sqrt(t)
	dx /= dl
	dy /= dl
	dz /= dl
	mz := a[2] - c[2]
	my := a[1] - c[1]
	mx := a[0] - c[0]
	bb := mx*dx + my*dy + mz*dz
	cc := mx*mx + my*my + mz*mz - r*r
	if cc > 0 && bb > 0 {
		return false
	}
	discr := bb*bb - cc
	if discr < 0 {
		return false
	}
	tResult := -bb - math.Sqrt(discr)

	if tResult*tResult > t {
		return false
	}
	if tResult < 0 {
		tResult = 0
	}
	return true
}

func IntersectPointAABB(a, min, max *Vector3) bool {
	for i := range 3 {
		if a[i] < min[i] {
			return false
		}
		if a[i] > max[i] {
			return false
		}
	}
	return true
}

func IntersectLineAABB(a, b, c, ext *Vector3) bool {
	d := Vector3{b[0] - a[0], b[1] - a[1], b[2] - a[2]}
	dl := d.Length()
	d[0] = dl / d[0]
	d[1] = dl / d[1]
	d[2] = dl / d[2]

	tx1 := (c[0] - ext[0]*0.5 - a[0]) * d[0]
	ty1 := (c[1] - ext[1]*0.5 - a[1]) * d[1]
	tz1 := (c[2] - ext[2]*0.5 - a[2]) * d[2]
	tx2 := (c[0] + ext[0]*0.5 - a[0]) * d[0]
	ty2 := (c[1] + ext[1]*0.5 - a[1]) * d[1]
	tz2 := (c[2] + ext[2]*0.5 - a[2]) * d[2]
	tmin := math.Max(math.Min(tx1, tx2), math.Min(ty1, ty2))
	tmin = math.Max(tmin, math.Min(tz1, tz2))
	tmax := math.Min(math.Max(tx1, tx2), math.Max(ty1, ty2))
	tmax = math.Min(tmax, math.Max(tz1, tz2))

	if tmax < 0 {
		return false
	}
	if tmin > tmax || tmin > dl {
		return false
	}
	return true
}

func IntersectSegmentsRaw(s1A, s1B, s2A, s2B *Vector2) (float64, float64, float64, float64) {
	s1dx := s1B[0] - s1A[0]
	s1dy := s1B[1] - s1A[1]
	s2dx := s2B[0] - s2A[0]
	s2dy := s2B[1] - s2A[1]

	denom := s1dx*s2dy - s2dx*s1dy
	if denom == 0 {
		return -1, -1, -1, -1
	}
	r := (s1A[1]-s2A[1])*s2dx - (s1A[0]-s2A[0])*s2dy
	if (denom < 0 && r >= constants.IntersectEpsilon) ||
		(denom > 0 && r < -constants.IntersectEpsilon) {
		return -1, -1, -1, -1
	}
	s := (s1A[1]-s2A[1])*s1dx - (s1A[0]-s2A[0])*s1dy
	if (denom < 0 && s >= constants.IntersectEpsilon) ||
		(denom > 0 && s < -constants.IntersectEpsilon) {
		return -1, -1, -1, -1
	}
	r /= denom
	s /= denom
	if r > 1.0+constants.IntersectEpsilon || s > 1.0+constants.IntersectEpsilon {
		return -1, -1, -1, -1
	}
	return Clamp(r, 0.0, 1.0), Clamp(s, 0.0, 1.0), s1dx, s1dy
}

func IntersectSegments(s1A, s1B, s2A, s2B, result *Vector2) bool {
	r, _, s1dx, s1dy := IntersectSegmentsRaw(s1A, s1B, s2A, s2B)
	if r < 0 {
		return false
	}
	result[0] = s1A[0] + r*s1dx
	result[1] = s1A[1] + r*s1dy
	return true
}

// Use the below like: `defer ExecutionDuration(ExecutionTrack("blah"))`
func ExecutionTrack(msg string) (string, time.Time) {
	return msg, time.Now()
}

func ExecutionDuration(msg string, start time.Time) {
	log.Printf("%v: %v\n", msg, time.Since(start))
}

// From https://gist.github.com/badboy/6267743
func Hash64to32(key uint64) uint32 {
	key = (^key) + (key << 18) // key = (key << 18) - key - 1;
	key = key ^ (key >> 31)
	key = key * 21 // key = (key + (key << 2)) + (key << 4);
	key = key ^ (key >> 11)
	key = key + (key << 6)
	key = key ^ (key >> 22)
	return uint32(key)
}
