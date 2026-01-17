// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"tlyakhov/gofoom/constants"
	"unsafe"
)

// Vector3 is a simple 3d vector type.
type Vector3 [3]float64

func (v *Vector3) To2D() *Vector2 {
	// In place cast works since it's just an array.
	return (*Vector2)(unsafe.Pointer(v))
}

// Zero returns true if all components are 0.
func (v *Vector3) Zero() bool {
	return v[0] == 0 && v[1] == 0 && v[2] == 0
}

// WithinEpsilon returns true if all components are within an epsilon.
func (v *Vector3) WithinEpsilon() bool {
	return v[0] > -constants.IntersectEpsilon && v[0] < constants.IntersectEpsilon &&
		v[1] > -constants.IntersectEpsilon && v[1] < constants.IntersectEpsilon &&
		v[2] > -constants.IntersectEpsilon && v[2] < constants.IntersectEpsilon
}

// Zero returns true if all components are within an epsilon.
func (v *Vector3) EqualEpsilon(v2 *Vector3) bool {
	return math.Abs(v2[0]-v[0]) < constants.IntersectEpsilon &&
		math.Abs(v2[1]-v[1]) < constants.IntersectEpsilon &&
		math.Abs(v2[2]-v[2]) < constants.IntersectEpsilon
}

func (v *Vector3) Clone() *Vector3 {
	return &Vector3{v[0], v[1], v[2]}
}

func (v *Vector3) From(v2 *Vector3) *Vector3 {
	v[0] = v2[0]
	v[1] = v2[1]
	v[2] = v2[2]
	return v
}

// Add a vector to a vector.
func (v *Vector3) Add(v2 *Vector3) *Vector3 {
	return &Vector3{v[0] + v2[0], v[1] + v2[1], v[2] + v2[2]}
}

// Add a vector to a vector.
func (v *Vector3) AddSelf(v2 *Vector3) *Vector3 {
	v[0] += v2[0]
	v[1] += v2[1]
	v[2] += v2[2]
	return v
}

// Sub subtracts a vector from a vector.
func (v *Vector3) Sub(v2 *Vector3) *Vector3 {
	return &Vector3{v[0] - v2[0], v[1] - v2[1], v[2] - v2[2]}
}

// Sub subtracts a vector from a vector.
func (v *Vector3) SubSelf(v2 *Vector3) *Vector3 {
	v[0] -= v2[0]
	v[1] -= v2[1]
	v[2] -= v2[2]
	return v
}

// Mul3 multiplies a vector by a vector.
func (v *Vector3) Mul3(v2 *Vector3) *Vector3 {
	return &Vector3{v[0] * v2[0], v[1] * v2[1], v[2] * v2[2]}
}

// Mul3 multiplies a vector by a vector.
func (v *Vector3) Mul3Self(v2 *Vector3) *Vector3 {
	v[0] *= v2[0]
	v[1] *= v2[1]
	v[2] *= v2[2]
	return v
}

// Mul multiplies a vector by a scalar.
func (v *Vector3) Mul(f float64) *Vector3 {
	return &Vector3{v[0] * f, v[1] * f, v[2] * f}
}

// Mul multiplies a vector by a scalar.
func (v *Vector3) MulSelf(f float64) *Vector3 {
	v[0] *= f
	v[1] *= f
	v[2] *= f
	return v
}

// Length calculates the length of a vector.
func (v *Vector3) Length() float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
}

// Length2 calculates the squared length of a vector.
func (v *Vector3) Length2() float64 {
	return v[0]*v[0] + v[1]*v[1] + v[2]*v[2]
}

// DistSq calculates the squared distance between two vectors.
func (v *Vector3) DistSq(v2 *Vector3) float64 {
	return (v[0]-v2[0])*(v[0]-v2[0]) +
		(v[1]-v2[1])*(v[1]-v2[1]) +
		(v[2]-v2[2])*(v[2]-v2[2])
}

// Dist calculates the distance between two vectors.
func (v *Vector3) Dist(v2 *Vector3) float64 {
	return math.Sqrt(v.DistSq(v2))
}

// Norm normalizes a vector and returns a new vector.
func (v *Vector3) Norm() *Vector3 {
	if v[0] == 0 && v[1] == 0 && v[2] == 0 {
		return &Vector3{0, 0, 0}
	}

	l := v.Length()
	return &Vector3{v[0] / l, v[1] / l, v[2] / l}
}

// NormSelf normalizes a vector in place.
func (v *Vector3) NormSelf() *Vector3 {
	if v[0] == 0 && v[1] == 0 && v[2] == 0 {
		return v
	}

	l := v.Length()

	v[0] /= l
	v[1] /= l
	v[2] /= l
	return v
}

// Dot calculates the dot product of two vectors.
func (a *Vector3) Dot(b *Vector3) float64 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

// Reflect reflects a vector around another vector.
func (v *Vector3) Reflect(normal *Vector3) *Vector3 {
	return (&Vector3{normal[0], normal[1], normal[2]}).MulSelf(2.0 * v.Dot(normal)).SubSelf(v)
}

// Reflect reflects a vector around another vector.
func (v *Vector3) ReflectSelf(normal *Vector3) *Vector3 {
	m := 2.0 * v.Dot(normal)
	v[0] = normal[0]*m - v[0]
	v[1] = normal[1]*m - v[1]
	v[2] = normal[2]*m - v[2]
	return v
}

// Clamp clamps a vector's values between a minimum and maximum range and returns a new vector.
func (v *Vector3) Clamp(min, max float64) *Vector3 {
	return &Vector3{Clamp(v[0], min, max), Clamp(v[1], min, max), Clamp(v[2], min, max)}
}

// ClampSelf clamps a vector's values between a minimum and maximum range in place.
func (v *Vector3) ClampSelf(min, max float64) *Vector3 {
	v[0] = Clamp(v[0], min, max)
	v[1] = Clamp(v[1], min, max)
	v[2] = Clamp(v[2], min, max)
	return v
}

// Deserialize assigns this vector's fields from a parsed JSON map.
func (v *Vector3) Deserialize(data string) {
	vx, _ := ParseVector3(data)
	*v = *vx
}

func (v *Vector3) ToInt32Color() uint32 {
	return uint32(v[0]*255)<<24 | uint32(v[1]*255)<<16 | uint32(v[2]*255)<<8 | 0xFF
}

// Cross computes the cross product of two vectors.
func (v *Vector3) Cross(vec2 *Vector3) *Vector3 {
	return &Vector3{v[1]*vec2[2] - v[2]*vec2[1], v[2]*vec2[0] - v[0]*vec2[2], v[0]*vec2[1] - v[1]*vec2[0]}
}

// CrossSelf computes the cross product of two vectors in place.
func (v *Vector3) CrossSelf(vec1, vec2 *Vector3) *Vector3 {
	v[0] = vec1[1]*vec2[2] - vec1[2]*vec2[1]
	v[1] = vec1[2]*vec2[0] - vec1[0]*vec2[2]
	v[2] = vec1[0]*vec2[1] - vec1[1]*vec2[0]
	return v
}

// From https://en.wikipedia.org/wiki/Spherical_coordinate_system#cartesian_coordinates
func (v *Vector3) ToSpherical() (r, theta, phi float64) {
	r = v[0]*v[0] + v[1]*v[1]
	phi = math.Atan2(v[1], v[0])
	r2d := math.Sqrt(r)
	r = math.Sqrt(r + v[2]*v[2])
	theta = math.Atan2(r2d, v[2])
	return
}

func (v *Vector3) FromSpherical(r, theta, phi float64) {
	v[2] = math.Cos(theta) * r
	v[1] = math.Sin(phi) * math.Sin(theta) * r
	v[0] = math.Cos(phi) * math.Sin(theta) * r
}

// Adapted from https://stackoverflow.com/questions/9038392/how-to-round-a-2d-vector-to-nearest-15-degrees?rq=4
func (v *Vector3) SphericalQuantIncPhi(phiSnap float64, multiples float64) {
	r2d := math.Sqrt(v[0]*v[0] + v[1]*v[1])
	angle := math.Atan2(v[1], v[0])
	phi := (math.Round(angle/phiSnap) + multiples) * phiSnap
	v[0] = math.Cos(phi) * r2d
	v[1] = math.Sin(phi) * r2d
}

func (v *Vector3) SphericalQuantIncTheta(thetaSnap float64, multiples float64) {
	r, theta, phi := v.ToSpherical()
	theta = (math.Round(theta/thetaSnap) + multiples) * thetaSnap
	v.FromSpherical(r, theta, phi)
}

// String formats the vector as a string
func (v *Vector3) String() string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(v[0], 'f', -1, 64) + ", " +
		strconv.FormatFloat(v[1], 'f', -1, 64) + ", " +
		strconv.FormatFloat(v[2], 'f', -1, 64)
}

// StringHuman formats the vector as a string with arbitrary precision.
func (v *Vector3) StringHuman(prec int) string {
	result := ""

	if math.Abs(v[0]) < constants.HumanQuantityEpsilon {
		result += "0, "
	} else {
		result += strconv.FormatFloat(v[0], 'f', prec, 64) + ", "
	}
	if math.Abs(v[1]) < constants.HumanQuantityEpsilon {
		result += "0, "
	} else {
		result += strconv.FormatFloat(v[1], 'f', prec, 64) + ", "
	}
	if math.Abs(v[2]) < constants.HumanQuantityEpsilon {
		result += "0"
	} else {
		result += strconv.FormatFloat(v[2], 'f', prec, 64)
	}
	return result
}

// Serialize formats the vector as a JSON key-value map.
func (v *Vector3) Serialize() string {
	return v.String()
}

// ParseVector3 parses strings in the form "X, Y, Z" into vectors.
func ParseVector3(s string) (*Vector3, error) {
	result := &Vector3{}
	split := strings.Split(s, ",")
	if len(split) != 3 {
		return nil, errors.New("can't parse Vector3: input string should have three comma-separated values")
	}
	var err error
	result[0], err = strconv.ParseFloat(strings.TrimSpace(split[0]), 64)
	if err != nil {
		return nil, err
	}
	result[1], err = strconv.ParseFloat(strings.TrimSpace(split[1]), 64)
	if err != nil {
		return nil, err
	}
	result[2], err = strconv.ParseFloat(strings.TrimSpace(split[2]), 64)
	if err != nil {
		return nil, err
	}
	return result, nil
}
