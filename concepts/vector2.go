package concepts

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/tlyakhov/gofoom/constants"
)

// Vector2 is a simple 2d vector type.
type Vector2 struct {
	X, Y float64
}

// Zero returns true if all components are 0.
func (v Vector2) Zero() bool {
	return v.X == 0 && v.Y == 0
}

// Add a vector to a vector.
func (v Vector2) Add(v2 Vector2) Vector2 {
	return Vector2{v.X + v2.X, v.Y + v2.Y}
}

// Sub subtracts a vector from a vector.
func (v Vector2) Sub(v2 Vector2) Vector2 {
	return Vector2{v.X - v2.X, v.Y - v2.Y}
}

// Mul2 multiplies a vector by a vector.
func (v Vector2) Mul2(v2 Vector2) Vector2 {
	return Vector2{v.X * v2.X, v.Y * v2.Y}
}

// Mul multiplies a vector by a scalar.
func (v Vector2) Mul(f float64) Vector2 {
	return Vector2{v.X * f, v.Y * f}
}

// Length calculates the length of a vector.
func (v Vector2) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// Dist2 calculates the squared distance between two vectors.
func (v Vector2) Dist2(v2 Vector2) float64 {
	return (v.X-v2.X)*(v.X-v2.X) +
		(v.Y-v2.Y)*(v.Y-v2.Y)
}

// Dist calculates the distance between two vectors.
func (v Vector2) Dist(v2 Vector2) float64 {
	return math.Sqrt(v.Dist2(v2))
}

// Norm normalizes a vector.
func (v Vector2) Norm() Vector2 {
	l := v.Length()

	if l == 0.0 {
		return Vector2{0, 0}
	}

	return Vector2{v.X / l, v.Y / l}
}

// Dot calculates the dot product of two vectors.
func (v Vector2) Dot(v2 Vector2) float64 {
	return v.X*v2.X + v.Y*v2.Y
}

// Reflect reflects a vector around another vector.
func (v Vector2) Reflect(normal Vector2) Vector2 {
	return normal.Mul(2.0 * v.Dot(normal)).Sub(v)
}

// Clamp clamps a vector's values between a minimum and maximum range.
func (v Vector2) Clamp(min, max float64) Vector2 {
	return Vector2{Clamp(v.X, min, max), Clamp(v.Y, min, max)}
}

// Floor gets the integer part of the vector's values.
func (v Vector2) Floor() Vector2 {
	return Vector2{math.Floor(v.X), math.Floor(v.Y)}
}

// Intersect returns the intersection of two 2D line segments.
func Intersect(s1A, s1B, s2A, s2B Vector2) (Vector2, bool) {
	s1dx := s1B.X - s1A.X
	s1dy := s1B.Y - s1A.Y
	s2dx := s2B.X - s2A.X
	s2dy := s2B.Y - s2A.Y

	denom := s1dx*s2dy - s2dx*s1dy
	if denom == 0 {
		return Vector2{}, false
	}
	r := (s1A.Y-s2A.Y)*s2dx - (s1A.X-s2A.X)*s2dy
	if (denom < 0 && r >= constants.IntersectEpsilon) ||
		(denom > 0 && r < -constants.IntersectEpsilon) {
		return Vector2{}, false
	}
	s := (s1A.Y-s2A.Y)*s1dx - (s1A.X-s2A.X)*s1dy
	if (denom < 0 && s >= constants.IntersectEpsilon) ||
		(denom > 0 && s < -constants.IntersectEpsilon) {
		return Vector2{}, false
	}
	r /= denom
	s /= denom
	if r < 0 || r > 1.0+constants.IntersectEpsilon || s > 1.0+constants.IntersectEpsilon {
		return Vector2{}, false
	}
	return Vector2{s1A.X + r*s1dx, s1A.Y + r*s1dy}, true
}

// To3D converts a 2D vector to 3D.
func (v Vector2) To3D() Vector3 {
	return Vector3{v.X, v.Y, 0}
}

// Deserialize assigns this vector's fields from a parsed JSON map.
func (v *Vector2) Deserialize(data map[string]interface{}) {
	if val, ok := data["X"]; ok {
		v.X = val.(float64)
	}
	if val, ok := data["Y"]; ok {
		v.Y = val.(float64)
	}
}

// String formats the vector as a string
func (v Vector2) String() string {
	return strconv.FormatFloat(v.X, 'f', -1, 64) + ", " +
		strconv.FormatFloat(v.Y, 'f', -1, 64)
}

// StringHuman formats the vector as a string with 2 digit precision.
func (v Vector2) StringHuman() string {
	return strconv.FormatFloat(v.X, 'f', 2, 64) + ", " +
		strconv.FormatFloat(v.Y, 'f', 2, 64)
}

// Serialize formats the vector as a JSON key-value map.
func (v Vector2) Serialize() map[string]interface{} {
	return map[string]interface{}{"X": v.X, "Y": v.Y}
}

// ParseVector2 parses strings in the form "X, Y" into vectors.
func ParseVector2(s string) (Vector2, error) {
	result := Vector2{}
	split := strings.Split(s, ",")
	if len(split) != 2 {
		return result, errors.New("can't parse Vector2: input string should have two comma-separated values")
	}
	var err error
	result.X, err = strconv.ParseFloat(strings.TrimSpace(split[0]), 64)
	if err != nil {
		return result, err
	}
	result.Y, err = strconv.ParseFloat(strings.TrimSpace(split[1]), 64)
	if err != nil {
		return result, err
	}
	return result, nil
}
