package concepts

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"time"
	"unicode"
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

func NearestPow2(n uint32) uint32 {
	// Invalid input
	if n < 1 {
		return 0
	}
	var res, curr uint32
	res = 1
	// Try all powers starting from 2^1
	for i := 0; i < 8*strconv.IntSize; i++ {
		curr = 1 << i
		// If current power is more than n, break
		if curr > n {
			break
		}
		res = curr
	}
	return res
	/*n--

	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16

	n++

	return n*/
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

func UMin(x, y uint32) uint32 {
	if x < y {
		return x
	}
	return y
}

func UMax(x, y uint32) uint32 {
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

func ColorToInt32(c color.Color) uint32 {
	r, g, b, a := c.RGBA()
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

var xorSeed uint64

func RngXorShift64() uint64 {
	if xorSeed == 0 {
		xorSeed = uint64(time.Now().UnixNano())
	}
	x := xorSeed
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	xorSeed = x
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
	delta := b.Sub(a)
	t := delta.Length2()
	delta.NormSelf()
	m := a.Sub(c)
	bb := m.Dot(delta)
	cc := m.Length2() - r*r
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
