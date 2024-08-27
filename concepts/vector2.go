// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"tlyakhov/gofoom/constants"
)

// Vector2 is a simple 2d vector type.
type Vector2 [2]float64

func V2(v *Vector2, x, y float64) *Vector2 {
	v[0] = x
	v[1] = y
	return v
}

// Zero returns true if all components are 0.
func (v *Vector2) Zero() bool {
	return v[0] == 0 && v[1] == 0
}

func (v *Vector2) From(v2 *Vector2) *Vector2 {
	v[0] = v2[0]
	v[1] = v2[1]
	return v
}

func (v *Vector2) Clone() *Vector2 {
	return &Vector2{v[0], v[1]}
}

// Add a vector to a vector.
func (v *Vector2) Add(v2 *Vector2) *Vector2 {
	return &Vector2{v[0] + v2[0], v[1] + v2[1]}
}

// Sub subtracts a vector from a vector.
func (v *Vector2) Sub(v2 *Vector2) *Vector2 {
	return &Vector2{v[0] - v2[0], v[1] - v2[1]}
}

// Mul2 multiplies a vector by a vector.
func (v *Vector2) Mul2(v2 *Vector2) *Vector2 {
	return &Vector2{v[0] * v2[0], v[1] * v2[1]}
}

// Mul multiplies a vector by a scalar.
func (v *Vector2) Mul(f float64) *Vector2 {
	return &Vector2{v[0] * f, v[1] * f}
}

// Add a vector to a vector.
func (v *Vector2) AddSelf(v2 *Vector2) *Vector2 {
	v[0] += v2[0]
	v[1] += v2[1]
	return v
}

// Sub subtracts a vector from a vector.
func (v *Vector2) SubSelf(v2 *Vector2) *Vector2 {
	v[0] -= v2[0]
	v[1] -= v2[1]
	return v
}

// Mul2 multiplies a vector by a vector.
func (v *Vector2) Mul2Self(v2 *Vector2) *Vector2 {
	v[0] *= v2[0]
	v[1] *= v2[1]
	return v
}

// Mul multiplies a vector by a scalar.
func (v *Vector2) MulSelf(f float64) *Vector2 {
	v[0] *= f
	v[1] *= f
	return v
}

// Length calculates the length of a vector.
func (v *Vector2) Length() float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1])
}

// Length2 calculates the length*length of a vector.
func (v *Vector2) Length2() float64 {
	return v[0]*v[0] + v[1]*v[1]
}

// Dist2 calculates the squared distance between two vectors.
func (v *Vector2) Dist2(v2 *Vector2) float64 {
	return (v[0]-v2[0])*(v[0]-v2[0]) +
		(v[1]-v2[1])*(v[1]-v2[1])
}

// Dist calculates the distance between two vectors.
func (v *Vector2) Dist(v2 *Vector2) float64 {
	return math.Sqrt(v.Dist2(v2))
}

// Norm normalizes a vector.
func (v *Vector2) Norm() *Vector2 {
	if v[0] == 0 && v[1] == 0 {
		return &Vector2{0, 0}
	}

	l := v.Length()
	return &Vector2{v[0] / l, v[1] / l}
}

// NormSelf normalizes a vector.
func (v *Vector2) NormSelf() *Vector2 {
	if v[0] == 0 && v[1] == 0 {
		return v
	}

	l := v.Length()
	v[0] /= l
	v[1] /= l

	return v
}

// Dot calculates the dot product of two vectors.
func (v *Vector2) Dot(v2 *Vector2) float64 {
	return v[0]*v2[0] + v[1]*v2[1]
}

// Cross calculates the cross product of two vectors.
func (v *Vector2) Cross(v2 *Vector2) float64 {
	return v[0]*v2[1] - v[1]*v2[0]
}

// Reflect reflects a vector around another vector.
func (v *Vector2) Reflect(normal *Vector2) *Vector2 {
	m := 2.0 * v.Dot(normal)
	return &Vector2{normal[0]*m - v[0], normal[1]*m - v[1]}
}

func (v *Vector2) AngularCross(amount float64) *Vector2 {
	return &Vector2{-amount * v[1], amount * v[0]}
}

// Reflect reflects a vector around another vector.
func (v *Vector2) ReflectSelf(normal *Vector2) *Vector2 {
	m := 2.0 * v.Dot(normal)
	v[0] = normal[0]*m - v[0]
	v[1] = normal[1]*m - v[1]
	return v
}

// Clamp clamps a vector's values between a minimum and maximum range.
func (v *Vector2) Clamp(min, max float64) *Vector2 {
	return &Vector2{Clamp(v[0], min, max), Clamp(v[1], min, max)}
}

// Clamp clamps a vector's values between a minimum and maximum range.
func (v *Vector2) ClampSelf(min, max float64) *Vector2 {
	v[0] = Clamp(v[0], min, max)
	v[1] = Clamp(v[1], min, max)
	return v
}

// Floor gets the integer part of the vector's values.
func (v *Vector2) Floor() *Vector2 {
	return &Vector2{math.Floor(v[0]), math.Floor(v[1])}
}

// Floor gets the integer part of the vector's values.
func (v *Vector2) FloorSelf() *Vector2 {
	v[0] = math.Floor(v[0])
	v[1] = math.Floor(v[1])
	return v
}

// Intersect returns the intersection of two 2D line segments.
func Intersect(s1A, s1B, s2A, s2B *Vector2) (*Vector2, bool) {
	s1dx := s1B[0] - s1A[0]
	s1dy := s1B[1] - s1A[1]
	s2dx := s2B[0] - s2A[0]
	s2dy := s2B[1] - s2A[1]

	denom := s1dx*s2dy - s2dx*s1dy
	if denom == 0 {
		return nil, false
	}
	r := (s1A[1]-s2A[1])*s2dx - (s1A[0]-s2A[0])*s2dy
	if (denom < 0 && r >= constants.IntersectEpsilon) ||
		(denom > 0 && r < -constants.IntersectEpsilon) {
		return nil, false
	}
	s := (s1A[1]-s2A[1])*s1dx - (s1A[0]-s2A[0])*s1dy
	if (denom < 0 && s >= constants.IntersectEpsilon) ||
		(denom > 0 && s < -constants.IntersectEpsilon) {
		return nil, false
	}
	r /= denom
	s /= denom
	if r < 0 || r > 1.0+constants.IntersectEpsilon || s > 1.0+constants.IntersectEpsilon {
		return nil, false
	}
	return &Vector2{s1A[0] + r*s1dx, s1A[1] + r*s1dy}, true
}

// To3D converts a 2D vector to 3D.
func (v *Vector2) To3D(v3 *Vector3) *Vector3 {
	v3[0] = v[0]
	v3[1] = v[1]
	v3[2] = 0
	return v3
}

// Deserialize assigns this vector's fields from a parsed JSON map.
func (v *Vector2) Deserialize(data map[string]any) {
	if val, ok := data["X"]; ok {
		v[0] = val.(float64)
	}
	if val, ok := data["Y"]; ok {
		v[1] = val.(float64)
	}
}

// String formats the vector as a string
func (v *Vector2) String() string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(v[0], 'f', -1, 64) + ", " +
		strconv.FormatFloat(v[1], 'f', -1, 64)
}

// StringHuman formats the vector as a string with 2 digit precision.
func (v *Vector2) StringHuman() string {
	return strconv.FormatFloat(v[0], 'f', 2, 64) + ", " +
		strconv.FormatFloat(v[1], 'f', 2, 64)
}

// Serialize formats the vector as a JSON key-value map.
func (v *Vector2) Serialize() map[string]any {
	return map[string]any{"X": v[0], "Y": v[1]}
}

// ParseVector2 parses strings in the form "X, Y" into vectors.
func ParseVector2(s string) (*Vector2, error) {
	result := &Vector2{}
	split := strings.Split(s, ",")
	if len(split) != 2 {
		return result, errors.New("can't parse Vector2: input string should have two comma-separated values")
	}
	var err error
	result[0], err = strconv.ParseFloat(strings.TrimSpace(split[0]), 64)
	if err != nil {
		return result, err
	}
	result[1], err = strconv.ParseFloat(strings.TrimSpace(split[1]), 64)
	if err != nil {
		return result, err
	}
	return result, nil
}

func Vector2AABBIntersect(amin, amax, bmin, bmax *Vector2, includeEdges bool) bool {
	if includeEdges {
		return (amin[0] <= bmax[0] &&
			amax[0] >= bmin[0] &&
			amin[1] <= bmax[1] &&
			amax[1] >= bmin[1])
	} else {
		return (amin[0] < bmax[0] &&
			amax[0] > bmin[0] &&
			amin[1] < bmax[1] &&
			amax[1] > bmin[1])
	}
}
