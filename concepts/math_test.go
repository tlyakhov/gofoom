// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts_test

import (
	"math/rand"
	"testing"
)

func BenchmarkByteClamp(b *testing.B) {
	b.Run("Branchless", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			byteClampBranchless(int(rand.Float64()*300 - 30))
		}
	})
	b.Run("Branching", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			byteClamp(rand.Float64()*300 - 30)
		}
	})
}

func byteClamp(x float64) uint8 {
	if x <= 0 {
		return 0
	}
	if x >= 0xFF {
		return 0xFF
	}

	return uint8(x)
}
func byteClampBranchless(x int) uint8 {
	// if x < 0, x = 0
	z := x & ^(x >> 63)
	return uint8(z ^ ((0xFF ^ z) & ((0xFF - z) >> 63)))
}
