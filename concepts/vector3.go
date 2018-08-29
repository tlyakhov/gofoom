package concepts

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

// Vector3 is a simple 3d vector type.
type Vector3 struct {
	X, Y, Z float64
}

// Zero returns true if all components are 0.
func (v Vector3) Zero() bool {
	return v.X == 0 && v.Y == 0 && v.Z == 0
}

// Add a vector to a vector.
func (v Vector3) Add(v2 Vector3) Vector3 {
	return Vector3{v.X + v2.X, v.Y + v2.Y, v.Z + v2.Z}
}

// Sub subtracts a vector from a vector.
func (v Vector3) Sub(v2 Vector3) Vector3 {
	return Vector3{v.X - v2.X, v.Y - v2.Y, v.Z - v2.Z}
}

// Mul3 multiplies a vector by a vector.
func (v Vector3) Mul3(v2 Vector3) Vector3 {
	return Vector3{v.X * v2.X, v.Y * v2.Y, v.Z * v2.Z}
}

// Mul multiplies a vector by a scalar.
func (v Vector3) Mul(f float64) Vector3 {
	return Vector3{v.X * f, v.Y * f, v.Z * f}
}

// Length calculates the length of a vector.
func (v Vector3) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// Dist2 calculates the squared distance between two vectors.
func (v Vector3) Dist2(v2 Vector3) float64 {
	return (v.X-v2.X)*(v.X-v2.X) +
		(v.Y-v2.Y)*(v.Y-v2.Y) +
		(v.Z-v2.Z)*(v.Z-v2.Z)
}

// Dist calculates the distance between two vectors.
func (v Vector3) Dist(v2 Vector3) float64 {
	return math.Sqrt(v.Dist2(v2))
}

// Norm normalizes a vector.
func (v Vector3) Norm() Vector3 {
	l := v.Length()

	if l == 0.0 {
		return Vector3{0, 0, 0}
	}

	return Vector3{v.X / l, v.Y / l, v.Z / l}
}

// Dot calculates the dot product of two vectors.
func (v Vector3) Dot(v2 Vector3) float64 {
	return v.X*v2.X + v.Y*v2.Y + v.Z*v2.Z
}

// Reflect reflects a vector around another vector.
func (v Vector3) Reflect(normal Vector3) Vector3 {
	return normal.Mul(2.0 * v.Dot(normal)).Sub(v)
}

// Clamp clamps a vector's values between a minimum and maximum range.
func (v Vector3) Clamp(min, max float64) Vector3 {
	return Vector3{Clamp(v.X, min, max), Clamp(v.Y, min, max), Clamp(v.Z, min, max)}
}

// To2D converts a 3D vector to 2D
func (v Vector3) To2D() Vector2 {
	return Vector2{v.X, v.Y}
}

// Deserialize assigns this vector's fields from a parsed JSON map.
func (v *Vector3) Deserialize(data map[string]interface{}) {
	if val, ok := data["X"]; ok {
		v.X = val.(float64)
	}
	if val, ok := data["Y"]; ok {
		v.Y = val.(float64)
	}
	if val, ok := data["Z"]; ok {
		v.Z = val.(float64)
	}
}

func (v Vector3) ToInt32Color() uint32 {
	return uint32(v.X)<<24 | uint32(v.Y)<<16 | uint32(v.Z)<<8 | 0xFF
}

// Cross computes the cross product of two vectors.
func (v Vector3) Cross(vec2 Vector3) Vector3 {
	return Vector3{v.Y*vec2.Z - v.Z*vec2.Y, v.Z*vec2.X - v.X*vec2.Z, v.X*vec2.Y - v.Y*vec2.X}
}

// String formats the vector as a string
func (v Vector3) String() string {
	return strconv.FormatFloat(v.X, 'f', -1, 64) + ", " +
		strconv.FormatFloat(v.Y, 'f', -1, 64) + ", " +
		strconv.FormatFloat(v.Z, 'f', -1, 64)
}

// StringHuman formats the vector as a string with 2 digit precision.
func (v Vector3) StringHuman() string {
	return strconv.FormatFloat(v.X, 'f', 2, 64) + ", " +
		strconv.FormatFloat(v.Y, 'f', 2, 64) + ", " +
		strconv.FormatFloat(v.Z, 'f', 2, 64)
}

// Serialize formats the vector as a JSON key-value map.
func (v Vector3) Serialize() map[string]interface{} {
	return map[string]interface{}{"X": v.X, "Y": v.Y, "Z": v.Z}
}

// ParseVector3 parses strings in the form "X, Y, Z" into vectors.
func ParseVector3(s string) (Vector3, error) {
	result := Vector3{}
	split := strings.Split(s, ",")
	if len(split) != 3 {
		return result, errors.New("can't parse Vector3: input string should have three comma-separated values")
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
	result.Z, err = strconv.ParseFloat(strings.TrimSpace(split[2]), 64)
	if err != nil {
		return result, err
	}
	return result, nil
}
