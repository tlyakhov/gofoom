// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"cmp"
	"fmt"
	"log"
	"math"
	"runtime"
	"strings"
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

func RngDecide(r uint64, mod uint64) bool {
	block := math.MaxUint64 / mod
	return r < block
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

// See https://tavianator.com/2015/ray_box_nan.html
// This seems to have a bug of some kind
func IntersectLineAABB(a, b, c, ext *Vector3) bool {
	d := Vector3{b[0] - a[0], b[1] - a[1], b[2] - a[2]}
	dl := 1.0 //d.Length()
	d[0] = dl / d[0]
	d[1] = dl / d[1]
	d[2] = dl / d[2]

	t1 := (c[0] - ext[0]*0.5 - a[0]) * d[0]
	t2 := (c[0] + ext[0]*0.5 - a[0]) * d[0]
	tmin := math.Min(t1, t2)
	tmax := math.Max(t1, t2)

	for i := 1; i < 3; i++ {
		t1 = (c[i] - ext[i]*0.5 - a[i]) * d[i]
		t2 = (c[i] + ext[i]*0.5 - a[i]) * d[i]
		tmin = math.Max(tmin, math.Min(math.Min(t1, t2), tmax))
		tmax = math.Min(tmax, math.Max(math.Max(t1, t2), tmin))
	}

	return tmax >= math.Max(tmin, 0.0)
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

func IntersectSegments(s1A, s1B, s2A, s2B, result *Vector2) float64 {
	r, _, s1dx, s1dy := IntersectSegmentsRaw(s1A, s1B, s2A, s2B)
	if r < 0 {
		return -1
	}
	result[0] = s1A[0] + r*s1dx
	result[1] = s1A[1] + r*s1dy
	return r
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

func StackTrace() string {
	var sb strings.Builder
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}
	return sb.String()
}
