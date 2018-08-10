package concepts

import (
	"math"
)

// Vector3 is a simple 3d vector type.
type Vector3 struct {
	X, Y, Z float64
}

var ZeroVector3 = Vector3{}

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
func (vec *Vector3) Deserialize(data map[string]interface{}) {
	if v, ok := data["X"]; ok {
		vec.X = v.(float64)
	}
	if v, ok := data["Y"]; ok {
		vec.Y = v.(float64)
	}
	if v, ok := data["Z"]; ok {
		vec.Z = v.(float64)
	}
}

func (vec Vector3) ToInt32Color() uint32 {
	return uint32(vec.X)<<24 | uint32(vec.Y)<<16 | uint32(vec.Z)<<8 | 0xFF
}
