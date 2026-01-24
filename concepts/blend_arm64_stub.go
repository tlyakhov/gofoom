//go:build arm64

package concepts

// blendFrameBuffer converts float64 framebuffer to uint8 with an added tint
//
//go:noescape
func blendFrameBuffer(target []uint8, fb [][4]float64, tint *[4]float64)

// blendColors adds a and b * opacity.
//
//go:noescape
func blendColors(a *[4]float64, b *[4]float64, opacity float64)

// AsmVector4Mul4Self multiplies a and b.
//
//go:noescape
func AsmVector4Mul4Self(a *[4]float64, b *[4]float64)

// AsmInt32ToVector4 converts a uint32 color to a vector.
//
//go:noescape
func AsmInt32ToVector4(c uint32, a *[4]float64)

// AsmInt32ToVector4PreMul converts a uint32 color to a vector with pre-multiplied alpha.
//
//go:noescape
func AsmInt32ToVector4PreMul(c uint32, a *[4]float64)
