// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"math"
	"tlyakhov/gofoom/constants"
)

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
