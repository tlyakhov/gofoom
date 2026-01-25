#include "textflag.h"

/*	TODO, vectorize this. See references here:
	https://github.com/Clement-Jean/simd-go/blob/main/arith_arm64.s
	https://eclecticlight.co/wp-content/uploads/2021/08/simdlanes.png?w=2048
	https://eclecticlight.co/2021/08/23/code-in-arm-assembly-lanes-and-loads-in-neon/
	https://developer.arm.com/documentation/dui0204/j/neon-and-vfp-programming/neon-load---store-element-and-structure-instructions/vldn-and-vstn--multiple-n-element-structures-
*/

// Constants
GLOBL ·data(SB), RODATA|NOPTR, $32
DATA ·data+0(SB)/8, $1.0
DATA ·data+8(SB)/8, $255.0
DATA ·data+16(SB)/8, $0.00392156862745098

// func blendFrameBuffer(target []uint8, fb [][4]float64, tint *[4]float64)
TEXT ·blendFrameBuffer(SB), NOSPLIT, $0-56
	MOVD target_base+0(FP), R0
	MOVD target_len+8(FP), R1
	MOVD fb_base+24(FP), R2
	MOVD tint+48(FP), R3

	CBZ R1, ret

	// Load 1.0 -> F20
	MOVD $·data+0(SB), R10
	FMOVD (R10), F20

	// Load 255.0 -> F21
	MOVD $·data+8(SB), R10
	FMOVD (R10), F21

	// Load tint -> F4, F5, F6, F7
	FMOVD (R3), F4  // R
	FMOVD 8(R3), F5 // G
	FMOVD 16(R3), F6 // B
	FMOVD 24(R3), F7 // A

	// Calc (1 - tint.a) -> F8
	FMOVD F20, F8
	FSUBD F7, F8 // F8 = 1.0 - A

	// Prepare 255 constant integer -> R14
	MOVD $255, R14

loop:
	// Load fb pixel -> F10, F11, F12, F13 (R, G, B, A)
	FMOVD (R2), F10
	FMOVD 8(R2), F11
	FMOVD 16(R2), F12
	FMOVD 24(R2), F13

	// fb * (1 - A)
	FMULD F8, F10
	FMULD F8, F11
	FMULD F8, F12
	FMULD F8, F13

	// + tint
	FADDD F4, F10
	FADDD F5, F11
	FADDD F6, F12
	FADDD F7, F13

	// * 255
	FMULD F21, F10
	FMULD F21, F11
	FMULD F21, F12
	FMULD F21, F13

	// Convert to Int64
	FCVTZSD F10, R10
	FCVTZSD F11, R11
	FCVTZSD F12, R12

	// Clamp R10 (R)
	CMP $0, R10
	CSEL LT, ZR, R10, R10
	CMP $255, R10
	CSEL GT, R14, R10, R10

	// Clamp R11 (G)
	CMP $0, R11
	CSEL LT, ZR, R11, R11
	CMP $255, R11
	CSEL GT, R14, R11, R11

	// Clamp R12 (B)
	CMP $0, R12
	CSEL LT, ZR, R12, R12
	CMP $255, R12
	CSEL GT, R14, R12, R12

	// Store bytes (R, G, B, 0xFF)
	MOVB R10, (R0)
	MOVB R11, 1(R0)
	MOVB R12, 2(R0)
	MOVB R14, 3(R0)

	// Advance
	ADD $4, R0
	ADD $32, R2
	SUB $4, R1
	CBNZ R1, loop

ret:
	RET

// func blendColors(a *[4]float64, b *[4]float64, opacity float64)
TEXT ·blendColors(SB), NOSPLIT, $0-24
	MOVD a+0(FP), R0
	MOVD b+8(FP), R1
	FMOVD opacity+16(FP), F0

	// Load 1.0 -> F20
	MOVD $·data+0(SB), R10
	FMOVD (R10), F20
	// Load 0.0 -> F21 (for clamping)
	FMOVD ZR, F21

	// Load b[3] -> F1
	FMOVD 24(R1), F1

	// inva = 1.0 - b[3] * o
	FMULD F0, F1
	FSUBD F1, F20, F2 // F2 = inva

	// Load a -> F10..F13
	FMOVD (R0), F10
	FMOVD 8(R0), F11
	FMOVD 16(R0), F12
	FMOVD 24(R0), F13

	// Load b -> F14..F17
	FMOVD (R1), F14
	FMOVD 8(R1), F15
	FMOVD 16(R1), F16
	FMOVD 24(R1), F17

	// a = a * inva
	FMULD F2, F10
	FMULD F2, F11
	FMULD F2, F12
	FMULD F2, F13

	// b = b * o (o is F0)
	FMULD F0, F14
	FMULD F0, F15
	FMULD F0, F16
	FMULD F0, F17

	// a = a + b
	FADDD F14, F10
	FADDD F15, F11
	FADDD F16, F12
	FADDD F17, F13

	// Clamp [0, 1]
	// 1.0 is F20. 0.0 is F21.

	// R
	FMAXD F21, F10
	FMIND F20, F10
	// G
	FMAXD F21, F11
	FMIND F20, F11
	// B
	FMAXD F21, F12
	FMIND F20, F12
	// A
	FMAXD F21, F13
	FMIND F20, F13

	// Store a
	FMOVD F10, (R0)
	FMOVD F11, 8(R0)
	FMOVD F12, 16(R0)
	FMOVD F13, 24(R0)

	RET

// func AsmVector4Mul4Self(a *[4]float64, b *[4]float64)
TEXT ·AsmVector4Mul4Self(SB), NOSPLIT, $0-16
	MOVD a+0(FP), R0
	MOVD b+8(FP), R1

	// Load a
	FMOVD (R0), F0
	FMOVD 8(R0), F1
	FMOVD 16(R0), F2
	FMOVD 24(R0), F3

	// Load b
	FMOVD (R1), F4
	FMOVD 8(R1), F5
	FMOVD 16(R1), F6
	FMOVD 24(R1), F7

	// Mul
	FMULD F4, F0
	FMULD F5, F1
	FMULD F6, F2
	FMULD F7, F3

	// Store a
	FMOVD F0, (R0)
	FMOVD F1, 8(R0)
	FMOVD F2, 16(R0)
	FMOVD F3, 24(R0)

	RET

// func AsmInt32ToVector4(c uint32, a *[4]float64)
TEXT ·AsmInt32ToVector4(SB), NOSPLIT, $0-16
	MOVWU c+0(FP), R0
	MOVD a+8(FP), R1

	// Load 1/255.0 -> F20
	MOVD $·data+16(SB), R10
	FMOVD (R10), F20

	// R: (c >> 24) & 0xFF
	LSR $24, R0, R2
	UCVTFD R2, F0
	FMULD F20, F0

	// G: (c >> 16) & 0xFF
	LSR $16, R0, R2
	AND $0xFF, R2, R2
	UCVTFD R2, F1
	FMULD F20, F1

	// B: (c >> 8) & 0xFF
	LSR $8, R0, R2
	AND $0xFF, R2, R2
	UCVTFD R2, F2
	FMULD F20, F2

	// A: c & 0xFF
	AND $0xFF, R0, R2
	UCVTFD R2, F3
	FMULD F20, F3

	// Store
	FMOVD F0, (R1)
	FMOVD F1, 8(R1)
	FMOVD F2, 16(R1)
	FMOVD F3, 24(R1)

	RET

// func AsmInt32ToVector4PreMul(c uint32, a *[4]float64)
TEXT ·AsmInt32ToVector4PreMul(SB), NOSPLIT, $0-16
	MOVWU c+0(FP), R0
	MOVD a+8(FP), R1

	// Load 1/255.0 -> F20
	MOVD $·data+16(SB), R10
	FMOVD (R10), F20

	// R
	LSR $24, R0, R2
	UCVTFD R2, F0
	FMULD F20, F0

	// G
	LSR $16, R0, R2
	AND $0xFF, R2, R2
	UCVTFD R2, F1
	FMULD F20, F1

	// B
	LSR $8, R0, R2
	AND $0xFF, R2, R2
	UCVTFD R2, F2
	FMULD F20, F2

	// A
	AND $0xFF, R0, R2
	UCVTFD R2, F3
	FMULD F20, F3

	// Pre-mul
	FMULD F3, F0
	FMULD F3, F1
	FMULD F3, F2

	// Store
	FMOVD F0, (R1)
	FMOVD F1, 8(R1)
	FMOVD F2, 16(R1)
	FMOVD F3, 24(R1)

	RET
