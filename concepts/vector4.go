package concepts

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"unsafe"
)

// Vector4 is a simple 4d vector type.
type Vector4 [4]float64

func V4(v *Vector4, x, y, z, w float64) *Vector4 {
	v[0] = x
	v[1] = y
	v[2] = z
	v[3] = w
	return v
}

func (v *Vector4) To3D() *Vector3 {
	// In place cast works since it's just an array.
	return (*Vector3)(unsafe.Pointer(v))
}
func (v *Vector4) To2D() *Vector2 {
	// In place cast works since it's just an array.
	return (*Vector2)(unsafe.Pointer(v))
}

// Zero returns true if all components are 0.
func (v *Vector4) Zero() bool {
	return v[0] == 0 && v[1] == 0 && v[2] == 0 && v[3] == 0
}

func (v *Vector4) Clone() *Vector4 {
	return &Vector4{v[0], v[1], v[2], v[3]}
}

func (v *Vector4) From(v2 *Vector4) *Vector4 {
	v[0] = v2[0]
	v[1] = v2[1]
	v[2] = v2[2]
	v[3] = v2[3]
	return v
}

// Add a vector to a vector.
func (a *Vector4) Add(b *Vector4) *Vector4 {
	return &Vector4{a[0] + b[0], a[1] + b[1], a[2] + b[2], a[3] + b[3]}
}

// Add a vector to a vector.
func (a *Vector4) AddSelf(b *Vector4) *Vector4 {
	a[0] += b[0]
	a[1] += b[1]
	a[2] += b[2]
	a[3] += b[3]
	return a
}

// Add a premul alpha color to self
func (a *Vector4) AddPreMulColorSelf(b *Vector4) *Vector4 {
	if b[3] == 0 {
		return a
	}
	inva := 1.0 - b[3]
	a[0] = a[0]*inva + b[0]
	a[1] = a[1]*inva + b[1]
	a[2] = a[2]*inva + b[2]
	a[3] = a[3]*inva + b[3]
	return a
}

// Sub subtracts a vector from a vector.
func (a *Vector4) Sub(b *Vector4) *Vector4 {
	return &Vector4{a[0] - b[0], a[1] - b[1], a[2] - b[2], a[3] - b[3]}
}

// Sub subtracts a vector from a vector.
func (a *Vector4) SubSelf(b *Vector4) *Vector4 {
	a[0] -= b[0]
	a[1] -= b[1]
	a[2] -= b[2]
	a[3] -= b[3]
	return a
}

// Mul3 multiplies a vector by a vector.
func (a *Vector4) Mul4(b *Vector4) *Vector4 {
	return &Vector4{a[0] * b[0], a[1] * b[1], a[2] * b[2], a[3] * b[3]}
}

// Mul3 multiplies a vector by a vector.
func (a *Vector4) Mul4Self(b *Vector4) *Vector4 {
	a[0] *= b[0]
	a[1] *= b[1]
	a[2] *= b[2]
	a[3] *= b[3]
	return a
}

// Mul multiplies a vector by a scalar.
func (v *Vector4) Mul(f float64) *Vector4 {
	return &Vector4{v[0] * f, v[1] * f, v[2] * f, v[3] * f}
}

// Mul multiplies a vector by a scalar.
func (v *Vector4) MulSelf(f float64) *Vector4 {
	v[0] *= f
	v[1] *= f
	v[2] *= f
	v[3] *= f
	return v
}

// Length calculates the length of a vector.
func (v *Vector4) Length() float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2] + v[3]*v[3])
}

// Length2 calculates the squared length of a vector.
func (v *Vector4) Length2() float64 {
	return v[0]*v[0] + v[1]*v[1] + v[2]*v[2] + v[3]*v[3]
}

// Dist2 calculates the squared distance between two vectors.
func (v *Vector4) Dist2(v2 *Vector4) float64 {
	return (v[0]-v2[0])*(v[0]-v2[0]) +
		(v[1]-v2[1])*(v[1]-v2[1]) +
		(v[2]-v2[2])*(v[2]-v2[2]) +
		(v[3]-v2[3])*(v[3]-v2[3])
}

// Dist calculates the distance between two vectors.
func (v *Vector4) Dist(v2 *Vector4) float64 {
	return math.Sqrt(v.Dist2(v2))
}

// Norm normalizes a vector and returns a new vector.
func (v *Vector4) Norm() *Vector4 {
	if v[0] == 0 && v[1] == 0 && v[2] == 0 && v[3] == 0 {
		return &Vector4{0, 0, 0}
	}

	l := v.Length()
	return &Vector4{v[0] / l, v[1] / l, v[2] / l, v[3] / l}
}

// NormSelf normalizes a vector in place.
func (v *Vector4) NormSelf() *Vector4 {
	if v[0] == 0 && v[1] == 0 && v[2] == 0 && v[3] == 0 {
		return v
	}

	l := v.Length()

	v[0] /= l
	v[1] /= l
	v[2] /= l
	v[3] /= l
	return v
}

// Dot calculates the dot product of two vectors.
func (a *Vector4) Dot(b *Vector4) float64 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] + a[3]*b[3]
}

// Clamp clamps a vector's values between a minimum and maximum range and returns a new vector.
func (v *Vector4) Clamp(min, max float64) *Vector4 {
	return &Vector4{
		Clamp(v[0], min, max), Clamp(v[1], min, max),
		Clamp(v[2], min, max), Clamp(v[3], min, max)}
}

// ClampSelf clamps a vector's values between a minimum and maximum range in place.
func (v *Vector4) ClampSelf(min, max float64) *Vector4 {
	v[0] = Clamp(v[0], min, max)
	v[1] = Clamp(v[1], min, max)
	v[2] = Clamp(v[2], min, max)
	v[3] = Clamp(v[3], min, max)
	return v
}

// Deserialize assigns this vector's fields from a parsed JSON map.
func (v *Vector4) Deserialize(data map[string]any, color bool) {
	if color {
		if val, ok := data["R"]; ok {
			v[0] = val.(float64)
		}
		if val, ok := data["G"]; ok {
			v[1] = val.(float64)
		}
		if val, ok := data["B"]; ok {
			v[2] = val.(float64)
		}
		if val, ok := data["A"]; ok {
			v[3] = val.(float64)
		}
	} else {
		if val, ok := data["X"]; ok {
			v[0] = val.(float64)
		}
		if val, ok := data["Y"]; ok {
			v[1] = val.(float64)
		}
		if val, ok := data["Z"]; ok {
			v[2] = val.(float64)
		}
		if val, ok := data["W"]; ok {
			v[3] = val.(float64)
		}
	}
}

func (v *Vector4) ToInt32Color() uint32 {
	return uint32(v[0]*255)<<24 | uint32(v[1]*255)<<16 | uint32(v[2]*255)<<8 | uint32(v[3]*255)
}

// String formats the vector as a string
func (v *Vector4) String() string {
	return strconv.FormatFloat(v[0], 'f', -1, 64) + ", " +
		strconv.FormatFloat(v[1], 'f', -1, 64) + ", " +
		strconv.FormatFloat(v[2], 'f', -1, 64) + ", " +
		strconv.FormatFloat(v[3], 'f', -1, 64)
}

// StringHuman formats the vector as a string with 2 digit precision.
func (v *Vector4) StringHuman() string {
	return strconv.FormatFloat(v[0], 'f', 2, 64) + ", " +
		strconv.FormatFloat(v[1], 'f', 2, 64) + ", " +
		strconv.FormatFloat(v[2], 'f', 2, 64) + ", " +
		strconv.FormatFloat(v[3], 'f', 2, 64)
}

// Serialize formats the vector as a JSON key-value map.
func (v *Vector4) Serialize(color bool) map[string]any {
	if color {
		return map[string]any{"R": v[0], "G": v[1], "B": v[2], "A": v[3]}
	}
	return map[string]any{"X": v[0], "Y": v[1], "Z": v[2], "W": v[3]}
}

// ParseVector4 parses strings in the form "X, Y, Z, W" into vectors.
func ParseVector4(s string) (*Vector4, error) {
	result := &Vector4{}
	split := strings.Split(s, ",")
	if len(split) != 4 {
		return nil, errors.New("can't parse Vector4: input string should have four comma-separated values")
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
	result[3], err = strconv.ParseFloat(strings.TrimSpace(split[3]), 64)
	if err != nil {
		return nil, err
	}
	return result, nil
}
